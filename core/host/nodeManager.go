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
	ch := make(chan nodeSession, 32)
	go mgr.process(ch)

	//initialize internal structures
	hosts, err := mgr.db.GetHosts()
	if err != nil {
		return nm{}, err
	}
	for _, h := range hosts {
		s := nodeSession{
			host:         h,
			hostSubs:     mgr.hostSubs,
			hostStatSubs: mgr.hostStatSubs,
			nodeMgr:      mgr,
			log:          log,
		}
		ch <- s
	}

	go mgr.listen(ch)

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

func (nm nm) process(newConns <-chan nodeSession) {
	conns := make(map[uint]nodeSession)

	//subscribe to own hosts to gather updates
	hostSub := nm.hostSubs.Subscribe()
	defer hostSub.Unsubscribe()

	for {
		select {
		case c := <-newConns:
			if con, ok := conns[c.host.ID]; ok {
				//record already exists, this is probably a new connection
				con.conn = c.conn
				con.cmdMsgs = make(chan Message, 32)
				go con.process()
				conns[c.host.ID] = con
			} else {
				conns[c.host.ID] = c
			}
		case u := <-hostSub.GetReceive():
			//host update from node, typically Running
			h := u.(mgm.Host)
			con, ok := conns[h.ID]
			if ok {
				con.host = h
				conns[h.ID] = con
			}
		case nc := <-nm.requestChan:
			switch nc.MessageType {
			case "StartRegion":
				if c, ok := conns[nc.Host.ID]; ok {
					c.cmdMsgs <- nc
				} else {
					nc.SR(false, "Host not found")
				}
			default:
				nc.SR(false, "Not Implemented")
			}

		case msg := <-nm.internalMsgs:
			switch msg.request {
			case "GetHosts":
				for _, c := range conns {
					msg.hosts <- c.host
				}
				close(msg.hosts)
			}
		}
	}
}

// NodeManager receives and communicates with mgm Node processes
func (nm nm) listen(newConns chan<- nodeSession) {

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

		s := nodeSession{host: host, conn: conn}
		newConns <- s
	}
}
