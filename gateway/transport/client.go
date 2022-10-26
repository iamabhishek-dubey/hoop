package transport

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/google/uuid"
	pb "github.com/runopsio/hoop/common/proto"
	"github.com/runopsio/hoop/gateway/client"
	"github.com/runopsio/hoop/gateway/connection"
	"github.com/runopsio/hoop/gateway/transport/plugins"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type (
	connectedClients struct {
		clients     map[string]pb.Transport_ConnectServer
		connections map[string]*connection.Connection
		mu          sync.Mutex
	}
)

var cc = connectedClients{
	clients:     make(map[string]pb.Transport_ConnectServer),
	connections: make(map[string]*connection.Connection),
	mu:          sync.Mutex{},
}

func bindClient(sessionID string, stream pb.Transport_ConnectServer, connection *connection.Connection) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.clients[sessionID] = stream
	cc.connections[sessionID] = connection
}

func unbindClient(id string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	delete(cc.clients, id)
	delete(cc.connections, id)
}

func getClientStream(id string) pb.Transport_ConnectServer {
	return cc.clients[id]
}

func (s *Server) subscribeClient(stream pb.Transport_ConnectServer, token string) error {
	ctx := stream.Context()
	md, _ := metadata.FromIncomingContext(ctx)

	hostname := extractData(md, "hostname")
	machineId := extractData(md, "machine_id")
	kernelVersion := extractData(md, "kernel_version")
	connectionName := extractData(md, "connection_name")

	sub, err := s.exchangeUserToken(token)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid authentication")
	}

	context, err := s.UserService.FindBySub(sub)
	if err != nil || context.User == nil {
		return status.Errorf(codes.Unauthenticated, "invalid authentication")
	}

	conn, err := s.ConnectionService.FindOne(context, connectionName)
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}

	if conn == nil {
		return status.Errorf(codes.NotFound, fmt.Sprintf("connection '%v' not found", connectionName))
	}

	sessionID := uuid.NewString()
	c := &client.Client{
		Id:            uuid.NewString(),
		SessionID:     sessionID,
		OrgId:         context.Org.Id,
		UserId:        context.User.Id,
		Hostname:      hostname,
		MachineId:     machineId,
		KernelVersion: kernelVersion,
		Status:        client.StatusConnected,
		ConnectionId:  conn.Id,
		AgentId:       conn.AgentId,
	}

	err = s.pluginOnConnectPhase(plugins.ParamsData{
		"session_id":      sessionID,
		"connection_id":   conn.Id,
		"connection_name": connectionName,
		"connection_type": string(conn.Type),
		"org_id":          context.Org.Id,
		"user_id":         context.User.Id,
		"hostname":        hostname,
		"machine_id":      machineId,
		"kernel_version":  kernelVersion,
	}, context)
	if err != nil {
		log.Printf("plugin refused to accept connection %q, err=%v", sessionID, err)
		return status.Errorf(codes.FailedPrecondition, err.Error())
	}

	s.ClientService.Persist(c)
	bindClient(c.SessionID, stream, conn)

	log.Printf("successful connection hostname: [%s], machineId [%s], kernelVersion [%s]", hostname, machineId, kernelVersion)
	clientErr := s.listenClientMessages(stream, c, conn)
	if err := s.pluginOnDisconnectPhase(plugins.ParamsData{
		"org_id":     context.Org.Id,
		"session_id": sessionID,
	}); err != nil {
		log.Printf("session=%v ua=client - failed processing plugin on-disconnect phase, err=%v", sessionID, err)
	}

	s.disconnectClient(c)
	return clientErr
}

func (s *Server) listenClientMessages(stream pb.Transport_ConnectServer, c *client.Client, conn *connection.Connection) error {
	ctx := stream.Context()

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// receive data from stream
		pkt, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			if status, ok := status.FromError(err); ok && status.Code() == codes.Canceled {
				// TODO: send packet to agent to clean up resources
				log.Printf("session=%v - client disconnected", c.SessionID)
				return nil
			}
			log.Printf("received error from client, err=%v", err)
			return status.Errorf(codes.Internal, "internal error, failed receiving client packet")
		}
		if pb.PacketType(pkt.Type) == pb.PacketKeepAliveType {
			continue
		}
		if pkt.Spec == nil {
			pkt.Spec = make(map[string][]byte)
		}
		pkt.Spec[pb.SpecGatewaySessionID] = []byte(c.SessionID)
		agStream := getAgentStream(conn.AgentId)
		if agStream == nil {
			log.Printf("agent connection not found for %q", c.AgentId)
			return status.Errorf(codes.FailedPrecondition, fmt.Sprintf("agent not found for %v", c.AgentId))
		}
		log.Printf("receive client packet type [%s] and session id [%s]", pkt.Type, c.SessionID)
		if err := s.pluginOnReceivePhase(c.SessionID, pkt); err != nil {
			log.Printf("plugin reject packet, err=%v", err)
			return status.Errorf(codes.Internal, "internal error, packet rejected, contact the administrator")
		}
		err = s.processClientPacket(pkt, c, conn, agStream)
		if err != nil {
			fmt.Printf("session=%v - failed processing client packet, err=%v", c.SessionID, err)
			return status.Errorf(codes.FailedPrecondition, "internal error, failed processing packet")
		}
	}
}

func (s *Server) processClientPacket(
	pkt *pb.Packet,
	client *client.Client,
	conn *connection.Connection,
	agentStream pb.Transport_ConnectServer) error {
	switch pb.PacketType(pkt.Type) {
	case pb.PacketGatewayConnectType:
		return s.processPacketGatewayConnect(pkt, client, conn, agentStream)
	default:
		_ = agentStream.Send(pkt)
	}
	return nil
}

func (s *Server) processPacketGatewayConnect(pkt *pb.Packet,
	client *client.Client,
	conn *connection.Connection,
	agentStream pb.Transport_ConnectServer) error {
	var clientArgs []string
	if pkt.Spec != nil {
		encArgs := pkt.Spec[pb.SpecClientExecArgsKey]
		if len(encArgs) > 0 {
			if err := pb.GobDecodeInto(encArgs, &clientArgs); err != nil {
				log.Printf("failed decoding args, err=%v", err)
			}
		}
	}
	encConnectionParams, err := pb.GobEncode(&pb.AgentConnectionParams{
		EnvVars:    conn.Secret,
		CmdList:    conn.Command,
		ClientArgs: clientArgs,
	})
	if err != nil {
		return fmt.Errorf("failed encoding connection params err=%v", err)
	}
	// TODO: deal with errors
	_ = agentStream.Send(&pb.Packet{
		Type: pb.PacketAgentConnectType.String(),
		Spec: map[string][]byte{
			pb.SpecGatewaySessionID:         []byte(client.SessionID),
			pb.SpecConnectionType:           []byte(conn.Type),
			pb.SpecAgentConnectionParamsKey: encConnectionParams,
		},
	})
	return nil
}

func (s *Server) disconnectClient(c *client.Client) {
	unbindClient(c.SessionID)
	c.Status = client.StatusDisconnected
	s.ClientService.Persist(c)
}

func (s *Server) exchangeUserToken(token string) (string, error) {
	if s.Profile == pb.DevProfile {
		return "test-user", nil
	}

	sub, err := s.IDProvider.VerifyAccessToken(token)
	if err != nil {
		return "", err
	}

	return sub, nil
}
