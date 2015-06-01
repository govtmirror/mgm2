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
	GetHosts() []mgm.Host
}

type regionManager interface {
	GetRegionsOnHost(mgm.Host) ([]mgm.Region, error)
}

// NewManager constructs NodeManager instances
func NewManager(port string, rMgr regionManager, db database.Database, log core.Logger) (Manager, error) {
	mgr := nm{}
	mgr.listenPort = port
	mgr.db = hostDatabase{db}
	mgr.logger = log
	mgr.hostSubs = core.NewSubscriptionManager()
	mgr.hostStatSubs = core.NewSubscriptionManager()
	mgr.internalMsgs = make(chan internalMsg, 32)
	mgr.requestChan = make(chan Message, 32)
	mgr.regionMgr = rMgr
	go mgr.listen()
	go mgr.process()

	//initialize internal cache
	hosts, err := mgr.db.GetHosts()
	if err != nil {
		return nm{}, err
	}
	for _, h := range hosts {
		mgr.hostSubs.Broadcast(h)
	}

	return mgr, nil
}

type nm struct {
	listenPort   string
	logger       core.Logger
	listener     net.Listener
	db           hostDatabase
	hostSubs     core.SubscriptionManager
	hostStatSubs core.SubscriptionManager
	regionMgr    regionManager

	requestChan  chan Message
	internalMsgs chan internalMsg
}

type internalMsg struct {
	request string
	hosts   chan mgm.Host
}

func (nm nm) GetHosts() []mgm.Host {
	var hosts []mgm.Host
	req := internalMsg{"GetHosts", make(chan mgm.Host, 32)}
	nm.internalMsgs <- req
	nm.logger.Info("Reading from host channel")
	for h := range req.hosts {
		hosts = append(hosts, h)
	}
	return hosts
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

	//subscribe to own hosts to gather updates
	hostSub := nm.hostSubs.Subscribe()
	defer hostSub.Unsubscribe()

	for {
		select {
		case u := <-hostSub.GetReceive():
			h := u.(mgm.Host)
			hosts[h.ID] = h
		case nc := <-nm.requestChan:
			nc.SR(false, "This part isnt here yet")

		case msg := <-nm.internalMsgs:
			switch msg.request {
			case "GetHosts":
				for _, h := range hosts {
					msg.hosts <- h
				}
				close(msg.hosts)
			}
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

		s := nodeSession{host, conn, nm.hostSubs, nm.hostStatSubs, nm.regionMgr, nm, nm.logger}
		go s.process()
	}
}

func (nm nm) connectionHandler(h mgm.Host, conn net.Conn) {

}
