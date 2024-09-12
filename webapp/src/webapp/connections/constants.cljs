(ns webapp.connections.constants
  (:require [clojure.string :as cs]))

(def connection-configs-required
  {:command-line []
   :tcp [{:key "host" :value ""}
         {:key "port" :value ""}]
   :mysql [{:key "host" :value "" :required true}
           {:key "user" :value "" :required true}
           {:key "pass" :value "" :required true}
           {:key "port" :value "" :required true}
           {:key "db" :value "" :required true}]
   :postgres [{:key "host" :value "" :required true}
              {:key "user" :value "" :required true}
              {:key "pass" :value "" :required true}
              {:key "port" :value "" :required true}
              {:key "db" :value "" :required true}
              {:key "sslmode" :value "" :required false}]
   :mssql [{:key "host" :value "" :required true}
           {:key "user" :value "" :required true}
           {:key "pass" :value "" :required true}
           {:key "port" :value "" :required true}
           {:key "db" :value "" :required true}
           {:key "insecure" :value "false" :required false}]
   :oracledb [{:key "host" :value "" :required true}
              {:key "user" :value "" :required true}
              {:key "pass" :value "" :required true}
              {:key "port" :value "" :required true}
              {:key "ld_library_path" :value "/opt/oracle/instantclient_19_24" :hidden true :required true}
              {:key "sid" :placeholder "SID or Service name" :value "" :required true}]
   :mongodb [{:key "connection_string"
              :value ""
              :required true
              :placeholder "mongodb+srv://root:<password>@devcluster.mwb5sun.mongodb.net/"}]})

(def connection-icons-name-dictionary
  {:dark {:postgres "/web-v1/images/connections-logos/postgres_logo.svg"
          :postgres-csv "/web-v1/images/connections-logos/postgres_logo.svg"
          :command-line "/web-v1/images/connections-logos/dark/custom_dark.svg"
          :custom "/web-v1/images/connections-logos/dark/custom_dark.svg"
          :tcp "/web-v1/images/connections-logos/dark/tcp_dark.svg"
          :mysql "/web-v1/images/connections-logos/dark/mysql_dark.png"
          :mysql-csv "/web-v1/images/connections-logos/dark/mysql_dark.png"
          :aws "/web-v1/images/connections-logos/aws_logo.svg"
          :bastion "/web-v1/images/connections-logos/bastion_logo.svg"
          :heroku "/web-v1/images/connections-logos/heroku_logo.svg"
          :nodejs "/web-v1/images/connections-logos/node_logo.svg"
          :python "/web-v1/images/connections-logos/python_logo.svg"
          :ruby-on-rails "/web-v1/images/connections-logos/dark/rails_dark.svg"
          :clojure "/web-v1/images/connections-logos/clojure_logo.svg"
          :kubernetes "/web-v1/images/connections-logos/k8s_logo.svg"
          :sql-server-csv "/web-v1/images/connections-logos/sql-server_logo.svg"
          :sql-server "/web-v1/images/connections-logos/sql-server_logo.svg"
          :oracledb "/web-v1/images/connections-logos/oracle_logo.svg"
          :mssql "/web-v1/images/connections-logos/sql-server_logo.svg"
          :mongodb "/web-v1/images/connections-logos/mongodb_logo.svg"}
   :light {:postgres "/web-v1/images/connections-logos/postgres_logo.svg"
           :postgres-csv "/web-v1/images/connections-logos/postgres_logo.svg"
           :command-line "/web-v1/images/connections-logos/command-line.svg"
           :custom "/web-v1/images/connections-logos/command-line.svg"
           :tcp "/web-v1/images/connections-logos/tcp_logo.svg"
           :mysql "/web-v1/images/connections-logos/mysql_logo.png"
           :mysql-csv "/web-v1/images/connections-logos/mysql_logo.png"
           :aws "/web-v1/images/connections-logos/aws_logo.svg"
           :bastion "/web-v1/images/connections-logos/bastion_logo.svg"
           :heroku "/web-v1/images/connections-logos/heroku_logo.svg"
           :nodejs "/web-v1/images/connections-logos/node_logo.svg"
           :python "/web-v1/images/connections-logos/python_logo.svg"
           :ruby-on-rails "/web-v1/images/connections-logos/rails_logo.svg"
           :clojure "/web-v1/images/connections-logos/clojure_logo.svg"
           :kubernetes "/web-v1/images/connections-logos/k8s_logo.svg"
           :sql-server-csv "/web-v1/images/connections-logos/sql-server_logo.svg"
           :sql-server "/web-v1/images/connections-logos/sql-server_logo.svg"
           :oracledb "/web-v1/images/connections-logos/oracle_logo.svg"
           :mssql "/web-v1/images/connections-logos/sql-server_logo.svg"
           :mongodb "/web-v1/images/connections-logos/mongodb_logo.svg"}})

(defn get-connection-icon [connection & [theme]]
  (let [connection-icons (get connection-icons-name-dictionary (or theme :light))]
    (cond
      (not (cs/blank? (:subtype connection))) (get connection-icons (keyword (:subtype connection)))
      (not (cs/blank? (:icon_name connection))) (get connection-icons (keyword (:icon_name connection)))
      :else (get connection-icons (keyword (:type connection))))))

(def connection-commands
  {"nodejs" "node"
   "clojure" "clj"
   "python" "python3"
   "ruby-on-rails" "rails runner -"
   "postgres" ""
   "mysql" ""
   "mssql" ""
   "mongodb" ""})

(def connection-postgres-demo
  {:name "postgres-demo"
   :type "database"
   :subtype "postgres"
   :access_mode_runbooks "enabled"
   :access_mode_exec "enabled"
   :access_mode_connect "enabled"
   :access_schema "enabled"
   :agent_id ""
   :reviewers []
   :redact_enabled false
   :redact_types []
   :secret {:envvar:DB "ZGVsbHN0b3Jl"
            :envvar:DBNAME "ZGVsbHN0b3Jl"
            :envvar:HOST "ZGVtby1wZy1kYi5jaDcwN3JuYWl6amcudXMtZWFzdC0xLnJkcy5hbWF6b25hd3MuY29t"
            :envvar:INSECURE "ZmFsc2U="
            :envvar:PASS "ZG9sbGFyLW1hbmdlci1jYXJvdXNlLUhFQVJURUQ="
            :envvar:PORT "NTQzMg=="
            :envvar:USER "ZGVtb3JlYWRvbmx5"}
   :command []})
