package host

import (
	"errors"
	"net"
	"strconv"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/database"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Manager is the interface to mgmNodes
type Manager interface {
	SubscribeHost() core.Subscription
	SubscribeHostStats() core.Subscription
	SubscribeRegionStats() core.Subscription
	StartRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	KillRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	GetHostByID(id int) (mgm.Host, bool, error)
	GetHosts() []mgm.Host
	RemoveHost(mgm.Host) error
}

// NewManager constructs NodeManager instances
func NewManager(port int, rMgr region.Manager, db database.Database, log logger.Log) (Manager, error) {
	mgr := nm{}
	mgr.listenPort = port
	mgr.db = hostDatabase{db}
	mgr.logger = logger.Wrap("HOST", log)
	mgr.hostSubs = core.NewSubscriptionManager()
	mgr.hostStatSubs = core.NewSubscriptionManager()
	mgr.regionStatSubs = core.NewSubscriptionManager()
	mgr.internalMsgs = make(chan internalMsg, 32)
	mgr.requestChan = make(chan Message, 32)
	mgr.regionMgr = rMgr
	ch := make(chan hostSession, 32)
	go mgr.process(ch)

	//initialize internal structures
	hosts, err := mgr.db.GetHosts()
	if err != nil {
		return nm{}, err
	}
	for _, h := range hosts {
		s := hostSession{
			host:           h,
			hostSubs:       mgr.hostSubs,
			hostStatSubs:   mgr.hostStatSubs,
			regionStatSubs: mgr.regionStatSubs,
			nodeMgr:        mgr,
			regionMgr:      rMgr,
			log:            logger.Wrap(strconv.Itoa(h.ID), mgr.logger),
		}
		ch <- s
	}

	go mgr.listen(ch)

	return mgr, nil
}

type nm struct {
	listenPort     int
	logger         logger.Log
	listener       net.Listener
	db             hostDatabase
	hostSubs       core.SubscriptionManager
	hostStatSubs   core.SubscriptionManager
	regionStatSubs core.SubscriptionManager
	regionMgr      region.Manager

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

func (nm nm) GetHostByID(id int) (mgm.Host, bool, error) {
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

func (nm nm) KillRegionOnHost(region mgm.Region, host mgm.Host, sr core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "KillRegion",
		Region:      region,
		Host:        host,
		SR:          sr,
	}
}

func (nm nm) RemoveHost(host mgm.Host) error {
	return errors.New("not implemented")
}

func (nm nm) SubscribeHost() core.Subscription {
	return nm.hostSubs.Subscribe()
}

func (nm nm) SubscribeHostStats() core.Subscription {
	return nm.hostStatSubs.Subscribe()
}

func (nm nm) SubscribeRegionStats() core.Subscription {
	return nm.regionStatSubs.Subscribe()
}

func (nm nm) process(newConns <-chan hostSession) {
	conns := make(map[int]hostSession)

	haltedHost := make(chan int, 16)

	//subscribe to own hosts to gather updates
	hostSub := nm.hostSubs.Subscribe()
	defer hostSub.Unsubscribe()

	for {
		select {
		case c := <-newConns:
			if con, ok := conns[c.host.ID]; ok {
				//record already exists, this is probably a new connection
				con.Running = true
				con.conn = c.conn
				con.cmdMsgs = make(chan Message, 32)
				go con.process(haltedHost)
				conns[c.host.ID] = con
			} else {
				conns[c.host.ID] = c
			}
		case id := <-haltedHost:
			//a connection went offline
			if con, ok := conns[id]; ok {
				con.Running = false
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
					if !c.Running {
						nc.SR(false, "Host is not running")
						continue
					}
					//trigger region to record config files
					cfgs, err := nm.regionMgr.ServeConfigs(nc.Region, nc.Host)
					if err != nil {
						nc.SR(false, "Error getting region configs")
						continue
					}
					nc.Configs = cfgs
					c.cmdMsgs <- nc
				} else {
					nm.logger.Info("Host %v not found", nc.Host.ID)
					nc.SR(false, "Host not found, or not assigned")
				}
			case "KillRegion":
				if c, ok := conns[nc.Host.ID]; ok {
					if !c.Running {
						nc.SR(false, "Host is not running")
						continue
					}
					c.cmdMsgs <- nc
				} else {
					nm.logger.Info("Host %v not found", nc.Host.ID)
					nc.SR(false, "Host not found, or not assigned")
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
func (nm nm) listen(newConns chan<- hostSession) {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(nm.listenPort))
	if err != nil {
		nm.logger.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	nm.listener = ln
	nm.logger.Info("Listening for mgm host instances on :%d", nm.listenPort)

	for {
		conn, err := nm.listener.Accept()
		if err != nil {
			nm.logger.Error("Error accepting connection: ", err)
			continue
		}
		//validate connection, and identify host
		addr := conn.RemoteAddr()
		address := addr.(*net.TCPAddr).IP.String()
		host, exists, err := nm.db.GetHostByAddress(address)
		if err != nil {
			nm.logger.Error("Error looking up mgm Node %v: %v", address, err)
			conn.Close()
			continue
		}
		if !exists {
			nm.logger.Error("mgm Node %v does not exist", address, err)
			conn.Close()
			continue
		}
		if host.Address != address {
			nm.logger.Info("mgmNode connection from unregistered address: ", address)
			continue
		}
		nm.logger.Info("MGM Node connection from: %v (%v)", host.ID, address)

		s := hostSession{host: host, conn: conn}
		newConns <- s
	}
}
