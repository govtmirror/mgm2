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
	"github.com/satori/go.uuid"
)

// Manager is the interface to mgmNodes
type Manager interface {
	StartRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	KillRegionOnHost(mgm.Region, mgm.Host, core.ServiceRequest)
	RemoveHost(mgm.Host, core.ServiceRequest)
	RemoveRegionFromHost(mgm.Region, mgm.Host, core.ServiceRequest)
	AddRegionToHost(mgm.Region, mgm.Host, core.ServiceRequest)
	UpdateRegion(mgm.Region, core.ServiceRequest)
	SetRegionEstate(mgm.Region, mgm.Estate, core.ServiceRequest)
	AddHost(mgm.Host, core.ServiceRequest)
}

// NewManager constructs NodeManager instances
func NewManager(port int, rMgr region.Manager, pers persist.MGMDB, log logger.Log) (Manager, error) {
	mgr := nm{}
	mgr.listenPort = port
	mgr.mgm = pers
	mgr.log = logger.Wrap("HOST", log)
	mgr.internalMsgs = make(chan internalMsg, 32)
	mgr.requestChan = make(chan Message, 32)
	mgr.regionMgr = rMgr
	ch := make(chan hostSession, 32)
	go mgr.process(ch)

	go mgr.listen(ch)

	return mgr, nil
}

type nm struct {
	listenPort int
	log        logger.Log
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

type registrationRequest struct {
	reg  Registration
	host mgm.Host
}

type regionCommand struct {
	cmd    string
	region mgm.Region
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

func (nm nm) RemoveHost(h mgm.Host, callback core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "RemoveHost",
		Host:        h,
		SR:          callback,
	}
}

func (nm nm) AddHost(h mgm.Host, callback core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "AddHost",
		Host:        h,
		SR:          callback,
	}
}

func (nm nm) RemoveRegionFromHost(r mgm.Region, h mgm.Host, callback core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "RemoveFromHost",
		Region:      r,
		Host:        h,
		SR:          callback,
	}
}

func (nm nm) AddRegionToHost(r mgm.Region, h mgm.Host, callback core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "AssignToHost",
		Region:      r,
		Host:        h,
		SR:          callback,
	}
}

func (nm nm) UpdateRegion(r mgm.Region, callback core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "UpdateRegion",
		Region:      r,
		SR:          callback,
	}
}

func (nm nm) SetRegionEstate(r mgm.Region, e mgm.Estate, c core.ServiceRequest) {
	nm.requestChan <- Message{
		MessageType: "SetEstate",
		Region:      r,
		Estate:      e,
		SR:          c,
	}
}

func (nm nm) process(newConns <-chan hostSession) {
	conns := make(map[int64]hostSession)

	haltedHost := make(chan int64, 16)
	regs := make(chan registrationRequest, 16)

	hStatChan := make(chan mgm.HostStat, 32)
	rStatChan := make(chan mgm.RegionStat, 64)

	regionStats := make(map[uuid.UUID]mgm.RegionStat)

	//initialize internal structures
	for _, h := range nm.mgm.GetHosts() {
		s := hostSession{
			host: h,
			log:  logger.Wrap(strconv.FormatInt(h.ID, 10), nm.log),
		}
		conns[h.ID] = s
	}

Processing:
	for {
		select {
		case c := <-newConns:
			if con, ok := conns[c.host.ID]; ok {
				//connection from remote host
				con.Running = true
				con.conn = c.conn
				con.cmdMsgs = make(chan Message, 32)
				go con.process(haltedHost, regs, hStatChan, rStatChan)
				conns[c.host.ID] = con

				//make sure the host is populated with its regions
				for _, r := range nm.mgm.GetRegions() {
					if r.Host == con.host.ID {
						nc := Message{}
						nc.MessageType = "AddRegion"
						nc.Region = r
						nc.SR = func(bool, string) {}
						con.cmdMsgs <- nc
					}
				}

				//place host online
				c.host.Running = true
				nm.mgm.UpdateHost(c.host)
			} else {
				conns[c.host.ID] = c
			}
		case stat := <-hStatChan:
			nm.mgm.UpdateHostStat(stat)
		case stat := <-rStatChan:
			regionStats[stat.UUID] = stat
			nm.mgm.UpdateRegionStat(stat)
		case id := <-haltedHost:
			//a connection went offline
			if con, ok := conns[id]; ok {
				con.Running = false
				con.host.Running = false
				nm.mgm.UpdateHost(con.host)

				//offline regions on the disconnected host
				for _, reg := range nm.mgm.GetRegions() {
					if reg.Host == con.host.ID {
						if stat, ok := regionStats[reg.UUID]; ok {
							if stat.Running {
								stat.Running = false
								nm.mgm.UpdateRegionStat(stat)
							}
						}
					}
				}
			}
		case reg := <-regs:
			//recieved registration message for host, update record
			for _, c := range conns {
				if c.host.ID == reg.host.ID {
					//bingo
					c.host.ExternalAddress = reg.reg.ExternalAddress
					c.host.Hostname = reg.reg.Name
					c.host.Slots = reg.reg.Slots
					nm.mgm.UpdateHost(c.host)
				}
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
					nm.log.Info("Host %v not found", nc.Host.ID)
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
					nm.log.Info("Host %v not found", nc.Host.ID)
					nc.SR(false, "Host not found, or not assigned")
				}
			case "UpdateRegion":
				//regions can only be modified when they are halted
				go func() {
					for _, stat := range regionStats {
						if stat.UUID == nc.Region.UUID {
							if stat.Running {
								nc.SR(false, "Region cannot be updated while running")
								return
							}
						}
					}

					nm.mgm.UpdateRegion(nc.Region)

					nc.SR(true, "Region updated")

				}()

			case "SetEstate":
				//region-estate relationships can only be modified when the region is halted
				go func() {
					for _, stat := range regionStats {
						if stat.UUID == nc.Region.UUID {
							if stat.Running {
								nc.SR(false, "Region cannot be updated while running")
								return
							}
						}
					}

					nm.mgm.MoveRegionToEstate(nc.Region, nc.Estate)

					nc.SR(true, "Region updated")

				}()

			case "AssignToHost":
				//add a region to a host
				//make sure region is unassigned, and not running
				nm.log.Info("Adding region %v to host %v", nc.Region.UUID.String(), nc.Host.ID)

				for _, r := range nm.mgm.GetRegions() {
					if r.UUID == nc.Region.UUID && r.Host != 0 {
						nc.SR(false, "Error: The region is on a host")
						break Processing
					}
				}
				if stat, ok := regionStats[nc.Region.UUID]; ok {
					if stat.Running {
						nc.SR(false, "Error: The region is running")
						break Processing
					}
				}

				//persist change
				nc.Region.Host = nc.Host.ID
				nm.mgm.UpdateRegion(nc.Region)

				// notify host if active
				isRunning := false
				for _, con := range conns {
					if con.host.ID == nc.Host.ID {
						if con.Running {
							nc.MessageType = "AddRegion"
							con.cmdMsgs <- nc
							isRunning = true
						}
					}
				}

				if !isRunning {
					//host is down, handle callback now
					nc.SR(true, "Region Added to Host")
					nm.log.Info("Adding region %v to host %v Complete", nc.Region.UUID.String(), nc.Host.ID)
				}

			case "RemoveFromHost":
				//remove a region from a host
				//make sure region is on host in question, and not running
				nm.log.Info("Removing region %v from host %v", nc.Region.UUID.String(), nc.Host.ID)
				if nc.Region.Host == 0 {
					nc.SR(false, "Error: The region is not on a host")
					nm.log.Info("Removing region %v from host %v failed.  The region is not on a host.", nc.Region.UUID.String(), nc.Host.ID)
					break Processing
				}
				for _, r := range nm.mgm.GetRegions() {
					if r.UUID == nc.Region.UUID && r.Host != nc.Host.ID {
						nc.SR(false, "Error: The region is not on that host")
						nm.log.Info("Removing region %v from host %v failed.  The region is not on that host.", nc.Region.UUID.String(), nc.Host.ID)
						break Processing
					}
				}
				if stat, ok := regionStats[nc.Region.UUID]; ok {
					if stat.Running {
						nc.SR(false, "Error: The region is running")
						nm.log.Info("Removing region %v from host %v failed.  The region is currently running", nc.Region.UUID.String(), nc.Host.ID)
						break Processing
					}
				}

				//persist change
				nc.Region.Host = 0
				nm.mgm.UpdateRegion(nc.Region)

				// notify host if active
				isRunning := false
				for _, con := range conns {
					if con.host.ID == nc.Host.ID {
						if con.Running {
							nc.MessageType = "RemoveRegion"
							con.cmdMsgs <- nc
							isRunning = true
						}
					}
				}

				if !isRunning {
					//handle callback immediately
					nc.SR(true, "Region removed from Host")
					nm.log.Info("Removing region %v from host %v Complete", nc.Region.UUID.String(), nc.Host.ID)
				}

			case "RemoveHost":
				if c, ok := conns[nc.Host.ID]; ok {
					if c.Running {
						c.cmdMsgs <- nc
					}
					nm.mgm.RemoveHost(c.host)
					nc.SR(true, "Host Removed")
				} else {
					nc.SR(false, "Host Not Found")
				}
			case "AddHost":
				for _, h := range conns {
					if h.host.Address == nc.Host.Address {
						nc.SR(false, "Error, Host address already present")
					}
				}

				nm.mgm.AddHost(nc.Host)

				nc.SR(true, "Host Added")

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
		nm.log.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	nm.listener = ln
	nm.log.Info("Listening for mgm host instances on :%d", nm.listenPort)

	for {
		conn, err := nm.listener.Accept()
		if err != nil {
			nm.log.Error("Error accepting connection: ", err)
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
			nm.log.Error(errmsg)
			conn.Close()
			continue
		}
		if host.Address != address {
			nm.log.Info("mgmNode connection from unregistered address: ", address)
			continue
		}
		nm.log.Info("MGM Node connection from: %v (%v)", host.ID, address)

		//lookup current assignment of regions
		var regs []mgm.Region
		for _, r := range nm.mgm.GetRegions() {
			if r.Host == host.ID {
				regs = append(regs, r)
			}
		}

		s := hostSession{host: host, conn: conn, regions: regs}
		newConns <- s
	}
}
