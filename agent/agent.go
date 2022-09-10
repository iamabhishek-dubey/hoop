package main

import (
	"context"
	pb "github.com/runopsio/hoop/proto"
	"log"
)

type (
	Agent struct {
		stream      pb.Transport_ConnectClient
		ctx         context.Context
		closeSignal chan bool
	}
)

func (a *Agent) listen() {
	for {
		msg, err := a.stream.Recv()
		if err != nil {
			log.Printf("%s", err.Error())
			close(a.closeSignal)
			return
		}
		log.Printf("receive request type [%s] from component [%s]", msg.Type, msg.Component)
	}
}
