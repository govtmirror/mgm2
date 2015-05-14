package core

import (
	"encoding/json"
	"net"
)

// NodeManager receives and communicates with mgm Node processes
func NodeManager(listenPort string, hStatsLink chan<- HostStats, db Database, logger Logger) {

	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		logger.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	logger.Info("Listening for mgmNode instances on :" + listenPort)

	go mgmConnectionAcceptor(ln, hStatsLink, db, logger)
}

func mgmConnectionAcceptor(listen net.Listener, hStatsLink chan<- HostStats, db Database, logger Logger) {
	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Error("Error accepting connection: ", err)
			continue
		}
		//validate connection, and identify host
		addr := conn.RemoteAddr()
		address := addr.(*net.TCPAddr).IP.String()
		host, err := db.GetHostByAddress(address)
		if err != nil {
			logger.Error("Error looking up mgm Node: ", err)
			continue
		}
		if host.Address != address {
			logger.Info("mgmNode connection from unregistered address: ", address)
			continue
		}
		logger.Info("MGM Node connection from: %v", address)
		go mgmConnectionHandler(host.ID, conn, hStatsLink, logger)
	}
}

func mgmConnectionHandler(id uint, conn net.Conn, hStatsLink chan<- HostStats, logger Logger) {
	d := json.NewDecoder(conn)
	for {
		nmsg := NetworkMessage{}
		err := d.Decode(&nmsg)
		if err != nil {
			logger.Error("Error decoding mgmNode message: ", err)
			if err.Error() == "EOF" {
				return
			}
		}

		switch nmsg.MessageType {
		case "host_stats":
			hStats := nmsg.HStats
			hStats.ID = id
			hStatsLink <- hStats
		default:
			logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
		}

	}
}
