package core

import (
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
	for {
		data := make([]byte, 512)
		_, err := conn.Read(data)
		if err != nil {
			logger.Error("Error reading from socket: ", err)
			return
		}
		logger.Info("Received message from an MGM node: ", string(data))
	}
}
