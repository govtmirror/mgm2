package core

import (
	"encoding/json"
	"net"
)

// NodeManager is the interface to mgmNodes
type NodeManager interface {
}

// NewNodeManager constructs NodeManager instances
func NewNodeManager() NodeManager {
	mgr := nm{}
	go mgr.listen()
	return mgr
}

type nm struct {
	listenPort string
	logger     Logger
	listener   net.Listener
	db         Database
}

// NodeManager receives and communicates with mgm Node processes
func (nm nm) listen() {

	ln, err := net.Listen("tcp", ":"+nm.listenPort)
	if err != nil {
		nm.logger.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	nm.listener = ln
	nm.logger.Info("Listening for mgmNode instances on :" + nm.listenPort)

	for {
		conn, err := nm.listener.Accept()
		if err != nil {
			nm.logger.Error("Error accepting connection: ", err)
			continue
		}
		//validate connection, and identify host
		addr := conn.RemoteAddr()
		address := addr.(*net.TCPAddr).IP.String()
		host, err := nm.db.GetHostByAddress(address)
		if err != nil {
			nm.logger.Error("Error looking up mgm Node: ", err)
			continue
		}
		if host.Address != address {
			nm.logger.Info("mgmNode connection from unregistered address: ", address)
			continue
		}
		nm.logger.Info("MGM Node connection from: %v", address)
		go nm.connectionHandler(host.ID, conn, nm.db, nm.logger)
	}
}

func (nm nm) connectionHandler(id uint, conn net.Conn) {
	d := json.NewDecoder(conn)
	//place host online
	h, err := nm.db.PlaceHostOnline(id)
	if err != nil {
		nm.logger.Error("Error looking up host: ", err)
		return
	}
	hHub.HostNotifier <- h
	for {
		nmsg := NetworkMessage{}
		err := d.Decode(&nmsg)
		if err != nil {
			nm.logger.Error("Error decoding mgmNode message: ", err)
			if err.Error() == "EOF" {
				//place host offline
				h, err := nm.db.PlaceHostOffline(id)
				if err != nil {
					nm.logger.Error("Error looking up host: ", err)
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
			nm.logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
		}

	}
}
