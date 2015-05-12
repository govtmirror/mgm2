package core

import (
	"encoding/json"
	"net"
)

// NodeManager receives and communicates with mgm Node processes
func NodeManager(listenPort string, logger Logger) {

	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		logger.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	logger.Info("Listening for mgmNode instances on :" + listenPort)

	//inBox := make(chan []byte, 64)

	go mgmConnectionAcceptor(ln, logger)

	go func() {
		logger.Info("NodeManager Running")
	}()
}

func mgmConnectionAcceptor(listen net.Listener, logger Logger) {
	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Error("Error accepting connection: ", err)
			continue
		}
		go mgmConnectionHandler(conn, logger)
	}
}

func mgmConnectionHandler(conn net.Conn, logger Logger) {
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
			logger.Info("Received message from an MGM node: ", hStats)
		default:
			logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
		}

	}
}
