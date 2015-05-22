package core

import (
	"encoding/json"
	"net"
)

// NodeManager is the interface to mgmNodes
type NodeManager interface {
	SubscribeHost() subscription
	SubscribeHostStats() subscription
}

// NewNodeManager constructs NodeManager instances
func NewNodeManager(port string, db Database, log Logger) NodeManager {
	mgr := nm{}
	mgr.listenPort = port
	mgr.db = db
	mgr.logger = log
	mgr.hostSubs = newSubscriptionManager()
	mgr.hostStatSubs = newSubscriptionManager()
	go mgr.listen()
	return mgr
}

type nm struct {
	listenPort   string
	logger       Logger
	listener     net.Listener
	db           Database
	hostSubs     subscriptionManager
	hostStatSubs subscriptionManager
}

func (nm nm) SubscribeHost() subscription {
	return nm.hostSubs.Subscribe()
}

func (nm nm) SubscribeHostStats() subscription {
	return nm.hostStatSubs.Subscribe()
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
		nm.logger.Info("MGM Node connection from: %v (%v)", host.ID, address)
		go nm.connectionHandler(host.ID, conn)
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
	nm.hostSubs.Broadcast(h)

	for {
		nmsg := NetworkMessage{}
		err := d.Decode(&nmsg)
		if err != nil {
			if err.Error() == "EOF" {
				//place host offline
				h, err := nm.db.PlaceHostOffline(id)
				if err != nil {
					nm.logger.Error("Error looking up host: ", err)
					return
				}
				nm.hostSubs.Broadcast(h)
				nm.logger.Info("mgm node disconnected")
				return
			}
			nm.logger.Error("Error decoding mgmNode message: ", err)
		}

		switch nmsg.MessageType {
		case "host_stats":
			hStats := nmsg.HStats
			hStats.ID = id
			nm.hostStatSubs.Broadcast(hStats)
		default:
			nm.logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
		}

	}

}
