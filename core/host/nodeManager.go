package host

import (
	"net"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Manager is the interface to mgmNodes
type Manager interface {
	SubscribeHost() core.Subscription
	SubscribeHostStats() core.Subscription
	StartRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	GetHostByID(id uint) (mgm.Host, error)
	GetHosts() ([]mgm.Host, error)
}

// NewManager constructs NodeManager instances
func NewManager(port string, db database.Database, log core.Logger) Manager {
	mgr := nm{}
	mgr.listenPort = port
	mgr.db = hostDatabase{db}
	mgr.logger = log
	mgr.hostSubs = core.NewSubscriptionManager()
	mgr.hostStatSubs = core.NewSubscriptionManager()
	mgr.hostChan = make(chan mgm.Host, 16)
	mgr.requestChan = make(chan Message, 32)
	go mgr.listen()
	go mgr.process()
	return mgr
}

type nm struct {
	listenPort   string
	logger       core.Logger
	listener     net.Listener
	db           hostDatabase
	hostSubs     core.SubscriptionManager
	hostStatSubs core.SubscriptionManager

	hostChan    chan mgm.Host
	requestChan chan Message
}

func (nm nm) GetHosts() ([]mgm.Host, error) {
	return nm.db.GetHosts()
}

func (nm nm) GetHostByID(id uint) (mgm.Host, error) {
	return nm.db.GetHostByID(id)
}

func (nm nm) StartRegionOnHost(region mgm.Region, host mgm.Host, sr core.ServiceRequest) {
	nm.requestChan <- Message{
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
	h.Running = true
	nm.hostSubs.Broadcast(h)

	readMsgs := make(chan Message, 32)
	writeMsgs := make(chan Message, 32)
	nc := Comms{
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
			h.Running = false
			nm.hostSubs.Broadcast(h)
			return
		case nmsg := <-readMsgs:
			switch nmsg.MessageType {
			case "Register":
				reg := nmsg.Register
				h, err := nm.db.UpdateHost(h, reg)
				if err != nil {
					nm.logger.Error("Error registering new host: ", err.Error())
				}
				nm.hostSubs.Broadcast(h)
			case "HostStats":
				hStats := nmsg.HStats
				hStats.ID = h.ID
				nm.hostStatSubs.Broadcast(hStats)
			case "GetRegions":
				nm.logger.Info("Host %v requesting regions list: ", h.ID)
				regions, err := nm.db.GetRegionsOnHost(h)
				if err != nil {
					nm.logger.Error("Error getting regions for host: ", err.Error())
				} else {
					nm.logger.Info("Serving %v regions to Host %v", len(regions), h.ID)
					for _, r := range regions {
						writeMsgs <- Message{MessageType: "AddRegion", Region: r}
					}
				}
				nm.logger.Info("Region list served to Host %v", h.ID)
			default:
				nm.logger.Info("Received invalid message from an MGM node: ", nmsg.MessageType)
			}
		}

	}

}
