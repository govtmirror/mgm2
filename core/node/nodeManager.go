package node

import (
	"encoding/json"
	"net"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Manager is the interface to mgmNodes
type Manager interface {
	SubscribeHost() core.Subscription
	SubscribeHostStats() core.Subscription
	StartRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
}

// NewManager constructs NodeManager instances
func NewManager(port string, db core.Database, log core.Logger) Manager {
	mgr := nm{}
	mgr.listenPort = port
	mgr.db = db
	mgr.logger = log
	mgr.hostSubs = core.NewSubscriptionManager()
	mgr.hostStatSubs = core.NewSubscriptionManager()
	mgr.hostChan = make(chan mgm.Host, 16)
	mgr.requestChan = make(chan nodeControl, 32)
	go mgr.listen()
	go mgr.process()
	return mgr
}

type nm struct {
	listenPort   string
	logger       core.Logger
	listener     net.Listener
	db           core.Database
	hostSubs     core.SubscriptionManager
	hostStatSubs core.SubscriptionManager

	hostChan    chan mgm.Host
	requestChan chan nodeControl
}

type nodeControl struct {
	MessageType string
	Region      mgm.Region
	Host        mgm.Host
	SR          core.ServiceRequest
}

func (nm nm) StartRegionOnHost(region mgm.Region, host mgm.Host, sr core.ServiceRequest) {
	nm.requestChan <- nodeControl{
		MessageType: "StartRegion",
		Region:      region,
		Host:        host,
		SR:          sr,
	}
}

func (nm nm) SubscribeHost() core.Subscription {
	return nm.hostSubs.Subscribe()
}

func (nm nm) SubscribeHostStats() core.Subscription {
	return nm.hostStatSubs.Subscribe()
}

func (nm nm) process() {
	hosts := make(map[uint]mgm.Host)
	for {
		select {
		case host := <-nm.hostChan:
			hosts[host.ID] = host
		case nc := <-nm.requestChan:
			nc.SR(false, "This part isnt here yet")
		}
	}
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
		go nm.connectionHandler(host, conn)
	}
}

func (nm nm) connectionHandler(h mgm.Host, conn net.Conn) {
	//place host online
	h, err := nm.db.PlaceHostOnline(h.ID)
	if err != nil {
		nm.logger.Error("Error looking up host: ", err)
		return
	}
	nm.hostSubs.Broadcast(h)

	readMsgs := make(chan core.NetworkMessage, 32)
	writeMsgs := make(chan core.NetworkMessage, 32)
	nc := NodeConns{
		Connection: conn,
		Closing:    make(chan bool),
		Log:        nm.logger,
	}
	go nc.ReadConnection(readMsgs)
	go nc.WriteConnection(writeMsgs)

	for {

		select {
		case <-nc.Closing:
			nm.logger.Info("mgm node disconnected")
			h, err := nm.db.PlaceHostOffline(h.ID)
			if err != nil {
				nm.logger.Error("Error looking up host: ", err)
				return
			}
			nm.hostSubs.Broadcast(h)
			return
		case nmsg := <-readMsgs:
			switch nmsg.MessageType {
			case "HostStats":
				hStats := nmsg.HStats
				hStats.ID = h.ID
				nm.hostStatSubs.Broadcast(hStats)
			case "GetRegions":
				regions, err := nm.db.GetRegionsOnHost(h)
				if err != nil {
					nm.logger.Error("Error getting regions for host: ", err.Error())
				} else {
					for _, r := range regions {
						writeMsgs <- core.NetworkMessage{MessageType: "AddRegion", Region: r}
					}
				}
			default:
				nm.logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
			}
		}

	}

}

// NodeConns is a structure of shared read/write functions between MGM and MGMNode
type NodeConns struct {
	Connection net.Conn
	Closing    chan bool
	Log        core.Logger
}

// ReadConnection is a processing loop for reading a socket and parsing messages
func (node NodeConns) ReadConnection(readMsgs chan<- core.NetworkMessage) {
	d := json.NewDecoder(node.Connection)

	for {
		nmsg := core.NetworkMessage{}
		err := d.Decode(&nmsg)
		if err != nil {
			if err.Error() == "EOF" {
				close(node.Closing)
				node.Connection.Close()
				return
			}
			node.Log.Error("Error decoding mgm message: ", err)
		}

		readMsgs <- nmsg
	}
}

// WriteConnection is a processing loop for json encoding messages to a socket
func (node NodeConns) WriteConnection(writeMsgs <-chan core.NetworkMessage) {

	for {
		select {
		case <-node.Closing:
			return
		case msg := <-writeMsgs:
			data, _ := json.Marshal(msg)
			_, err := node.Connection.Write(data)
			if err != nil {
				node.Connection.Close()
				return
			}
		}
	}
}
