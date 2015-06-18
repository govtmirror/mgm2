package host

import (
	"fmt"
	"net"
	"strconv"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Manager is the interface to mgmNodes
type Manager interface {
	StartRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	KillRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	RemoveHost(mgm.Host, core.ServiceRequest)
}

// NewManager constructs NodeManager instances
func NewManager(port int, rMgr region.Manager, pers persist.MGMDB, log logger.Log) (Manager, error) {
	mgr := nm{}
	mgr.listenPort = port
	mgr.mgm = pers
	mgr.logger = logger.Wrap("HOST", log)
	mgr.internalMsgs = make(chan internalMsg, 32)
	mgr.requestChan = make(chan Message, 32)
	mgr.regionMgr = rMgr
	ch := make(chan hostSession, 32)
	go mgr.process(ch)

	//initialize internal structures
	hosts := mgr.mgm.GetHosts()
	for _, h := range hosts {
		s := hostSession{
			host:      h,
			nodeMgr:   mgr,
			regionMgr: rMgr,
			log:       logger.Wrap(strconv.Itoa(h.ID), mgr.logger),
		}
		ch <- s
	}

	go mgr.listen(ch)

	return mgr, nil
}

type nm struct {
	listenPort int
	logger     logger.Log
	listener   net.Listener
	mgm        persist.MGMDB
	regionMgr  region.Manager

	requestChan  chan Message
	internalMsgs chan internalMsg
}

type internalMsg struct {
	request string
	hosts   chan mgm.Host
}

/*func (nm nm) GetHosts() []mgm.Host {
	var hosts []mgm.Host
	req := internalMsg{"GetHosts", make(chan mgm.Host, 32)}
	nm.internalMsgs <- req
	nm.logger.Info("Reading from host channel")
	for h := range req.hosts {
		hosts = append(hosts, h)
	}
	return hosts
}

func (nm nm) GetHostByID(id int) (mgm.Host, bool) {
	hosts := nm.mgm.GetHosts()
	for _, h := range hosts {
		if h.ID == id {
			return h, true
		}
	}
	return mgm.Host{}, false
}*/

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

func (nm nm) RemoveHost(h mgm.Host, callback core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "RemoveHost",
		Host:        h,
		SR:          callback,
	}
}

func (nm nm) process(newConns <-chan hostSession) {
	conns := make(map[int]hostSession)

	haltedHost := make(chan int, 16)

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
		//case u := <-hostSub.GetReceive():
		//	//host update from node, typically Running
		//	h := u.(mgm.Host)
		//	con, ok := conns[h.ID]
		//	if ok {
		//		con.host = h
		//		conns[h.ID] = con
		//	}
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
			case "RemoveHost":
				if c, ok := conns[nc.Host.ID]; ok {
					if c.Running {
						c.cmdMsgs <- nc
					}
					nm.mgm.RemoveHost(c.host)
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
		hosts := nm.mgm.GetHosts()
		var host mgm.Host
		exists := false
		for _, h := range hosts {
			if h.Address == address {
				exists = true
				host = h
			}
		}
		if !exists {
			errmsg := fmt.Sprintf("mgm Node %v does not exist", address)
			nm.logger.Error(errmsg)
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
