package core

import (
	"encoding/json"
	"net"
)

// NodeManager receives and communicates with mgm Node processes
func NodeManager(listenPort string, hHub HostHub, db Database, logger Logger) {

	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		logger.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	logger.Info("Listening for mgmNode instances on :" + listenPort)

	go mgmConnectionAcceptor(ln, hHub, db, logger)
}

func mgmConnectionAcceptor(listen net.Listener, hHub HostHub, db Database, logger Logger) {
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
		go mgmConnectionHandler(host.ID, conn, hHub, db, logger)
	}
}

func mgmConnectionHandler(id uint, conn net.Conn, hHub HostHub, db Database, logger Logger) {
	d := json.NewDecoder(conn)
	//place host online
	h, err := db.PlaceHostOnline(id)
	if err != nil {
		logger.Error("Error looking up host: ", err)
		return
	}
	hHub.HostNotifier <- h
	for {
		nmsg := NetworkMessage{}
		err := d.Decode(&nmsg)
		if err != nil {
			logger.Error("Error decoding mgmNode message: ", err)
			if err.Error() == "EOF" {
				//place host offline
				h, err := db.PlaceHostOffline(id)
				if err != nil {
					logger.Error("Error looking up host: ", err)
					return
				}
				hHub.HostNotifier <- h
				return
			}
		}

		switch nmsg.MessageType {
		case "host_stats":
			hStats := nmsg.HStats
			hStats.ID = id
			hHub.HostStatsNotifier <- hStats
		default:
			logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
		}

	}
}
