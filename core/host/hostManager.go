package host

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
)

type notifier interface {
	HostRemoved(int64)
	HostUpdated(mgm.Host)
}

type hostConn struct {
}

func (hc hostConn) Close() {

}

// NewManager constructs NodeManager instances
func NewManager(port int, rMgr region.Manager, pers persist.MGMDB, notify notifier, log logger.Log) Manager {
	mgr := Manager{}
	mgr.listenPort = port
	mgr.mgm = pers
	mgr.log = logger.Wrap("HOST", log)
	mgr.internalMsgs = make(chan internalMsg, 32)
	mgr.requestChan = make(chan Message, 32)
	mgr.rMgr = rMgr
	mgr.notify = notify
	//ch := make(chan hostSession, 32)

	//go mgr.listen(ch)

	regions := rMgr.GetRegions()
	mgr.hosts = make(map[int64]mgm.Host)
	mgr.hostStats = make(map[int64]mgm.HostStat)
	mgr.hostConnections = make(map[int64]hostConn)
	mgr.hMutex = &sync.Mutex{}
	mgr.hsMutex = &sync.Mutex{}
	mgr.hcMutex = &sync.Mutex{}
	for _, h := range pers.QueryHosts() {
		mgr.hosts[h.ID] = h
		mgr.hostStats[h.ID] = mgm.HostStat{ID: h.ID}
	}
	for _, r := range regions {
		h, ok := mgr.hosts[r.Host]
		if ok {
			h.Regions = append(h.Regions, r.UUID)
			mgr.hosts[h.ID] = h
		}
	}

	return mgr
}

// Manager is a central access point for Host operations
type Manager struct {
	listenPort      int
	log             logger.Log
	listener        net.Listener
	mgm             persist.MGMDB
	rMgr            region.Manager
	notify          notifier
	hosts           map[int64]mgm.Host
	hMutex          *sync.Mutex
	hostConnections map[int64]hostConn
	hcMutex         *sync.Mutex
	hostStats       map[int64]mgm.HostStat
	hsMutex         *sync.Mutex

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

// GetHosts get a slice of all regions from cache
func (m Manager) GetHosts() []mgm.Host {
	m.hMutex.Lock()
	defer m.hMutex.Unlock()
	t := []mgm.Host{}
	for _, r := range m.hosts {
		t = append(t, r)
	}
	return t
}

// GetHostStats get a slice of all region stats from cache
func (m Manager) GetHostStats() []mgm.HostStat {
	m.hsMutex.Lock()
	defer m.hsMutex.Unlock()
	t := []mgm.HostStat{}
	for _, r := range m.hostStats {
		t = append(t, r)
	}
	return t
}

// StartRegionOnHost requests a region to be started with a matching host
func (m Manager) StartRegionOnHost(region mgm.Region, host mgm.Host) error {
	ch := make(chan error)
	m.requestChan <- Message{
		MessageType: "StartRegion",
		Region:      region,
		Host:        host,
		response:    ch,
	}
	err, ok := <-ch
	if ok {
		return nil
	}
	return err
}

// KillRegionOnHost requests a region to be killed on a specified host
func (m Manager) KillRegionOnHost(region mgm.Region, host mgm.Host) error {
	ch := make(chan error)
	m.requestChan <- Message{
		MessageType: "KillRegion",
		Region:      region,
		Host:        host,
		response:    ch,
	}
	err, ok := <-ch
	if ok {
		return nil
	}
	return err
}

// RemoveHost removes a host registration from MGM
func (m Manager) RemoveHost(id int64) error {
	m.log.Info("Removing host %v", id)
	m.hMutex.Lock()
	m.hsMutex.Lock()
	m.hcMutex.Lock()
	defer m.hcMutex.Unlock()
	defer m.hsMutex.Unlock()
	defer m.hMutex.Unlock()

	// hosts cannot be removed if they have regions assigned
	for _, r := range m.rMgr.GetRegions() {
		if r.Host == id {
			return errors.New("Host has regions assigned")
		}
	}

	//purge record from mysql
	m.mgm.PurgeHost(id)

	//close any connection to the node
	if hc, ok := m.hostConnections[id]; ok {
		hc.Close()
	}

	//remove the host from the cache
	delete(m.hostConnections, id)
	delete(m.hostStats, id)
	delete(m.hosts, id)

	//notify any users of the change
	m.notify.HostRemoved(id)

	m.log.Info("Host %v removed", id)

	return nil
}

// AddHost creates a new host registration in MGM
func (m Manager) AddHost(address string) error {
	m.log.Info("Adding host at %v", address)
	m.hMutex.Lock()
	defer m.hMutex.Unlock()

	//host cannot collide with existing addresses
	for _, h := range m.hosts {
		if h.Address == address {
			return errors.New("There is already a host at that address")
		}
	}

	id, err := m.mgm.InsertHost(address)
	if err != nil {
		return err
	}

	m.log.Info("New host %v at %v", id, address)

	newHost := mgm.Host{}
	newHost.ID = id
	newHost.Address = address

	m.hosts[id] = newHost
	m.hostStats[id] = mgm.HostStat{ID: id}
	m.notify.HostUpdated(newHost)
	return nil
}

// RemoveRegionFromHost deregisters a region form a host
func (m Manager) RemoveRegionFromHost(r mgm.Region, h mgm.Host) error {
	ch := make(chan error)
	m.requestChan <- Message{
		MessageType: "RemoveFromHost",
		Region:      r,
		Host:        h,
		response:    ch,
	}
	err, ok := <-ch
	if ok {
		return nil
	}
	return err
}

// AddRegionToHost registers a region to a specified host
func (m Manager) AddRegionToHost(r mgm.Region, h mgm.Host) error {
	ch := make(chan error)
	m.requestChan <- Message{
		MessageType: "AssignToHost",
		Region:      r,
		Host:        h,
		response:    ch,
	}
	err, ok := <-ch
	if ok {
		return nil
	}
	return err
}

// UpdateRegion modifies a region
func (m Manager) UpdateRegion(r mgm.Region) error {
	ch := make(chan error)
	m.requestChan <- Message{
		MessageType: "UpdateRegion",
		Region:      r,
		response:    ch,
	}
	err, ok := <-ch
	if ok {
		return nil
	}
	return err
}

// SetRegionEstate modifies a region
func (m Manager) SetRegionEstate(r mgm.Region, e mgm.Estate) error {
	ch := make(chan error)
	m.requestChan <- Message{
		MessageType: "SetEstate",
		Region:      r,
		Estate:      e,
		response:    ch,
	}
	err, ok := <-ch
	if ok {
		return nil
	}
	return err
}

/*
func (m Manager) process(newConns <-chan hostSession) {
	conns := make(map[int64]hostSession)

	haltedHost := make(chan int64, 16)
	regs := make(chan registrationRequest, 16)

	hStatChan := make(chan mgm.HostStat, 32)
	rStatChan := make(chan mgm.RegionStat, 64)

	//initialize internal structures
	for _, h := range m.mgm.QueryHosts() {
		s := hostSession{
			host: h,
			log:  logger.Wrap(strconv.FormatInt(h.ID, 10), m.log),
		}
		conns[h.ID] = s
	}

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
				for _, r := range m.rMgr.GetRegions() {
					if r.Host == con.host.ID {
						nc := Message{}
						nc.MessageType = "AddRegion"
						nc.Region = r
						con.cmdMsgs <- nc
					}
				}

				//place host online
				//st, ok := nm.mgm.GetHostStat(c.host.ID)
				//if ok {
				//	st.Running = true
				//	nm.mgm.UpdateHostStat(st)
				//} else {
				//	st.ID = c.host.ID
				//	st.Running = true
				//	nm.mgm.UpdateHostStat(st)
				//}
			} else {
				conns[c.host.ID] = c
			}
			//case stat := <-hStatChan:
			//	nm.mgm.UpdateHostStat(stat)
			//case stat := <-rStatChan:
			//	nm.mgm.UpdateRegionStat(stat)
			//case id := <-haltedHost:
			//a connection went offline
			if con, ok := conns[id]; ok {
				con.Running = false

				//offline the host
				st, _ := nm.mgm.GetHostStat(con.host.ID)
				st.Running = false
				nm.mgm.UpdateHostStat(st)

				//offline regions on the disconnected host
				for _, reg := range nm.mgm.GetRegions() {
					if reg.Host == con.host.ID {
						st, ok := nm.mgm.GetRegionStat(reg.UUID)
						if ok {
							st.Running = false
							nm.mgm.UpdateRegionStat(st)
						}
					}
				}
			}
*/
//case reg := <-regs:
/*	//recieved registration message for host, update record
	for _, c := range conns {
		if c.host.ID == reg.host.ID {
			//bingo
			c.host.ExternalAddress = reg.reg.ExternalAddress
			c.host.Hostname = reg.reg.Name
			c.host.Slots = reg.reg.Slots
			nm.mgm.UpdateHost(c.host)
		}
	}
*/
//case nc := <-nm.requestChan:
/*
	switch nc.MessageType {
	case "StartRegion":
		if c, ok := conns[nc.Host.ID]; ok {
			if !c.Running {
				nc.SR(false, "Host is not running")
				continue
			}
			//trigger region to record config files
			cfgs := nm.regionMgr.ServeConfigs(nc.Region, nc.Host)
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
			stat, ok := nm.mgm.GetRegionStat(nc.Region.UUID)
			if !ok {
				nc.SR(false, "Region not found")
				return
			}
			if stat.Running {
				nc.SR(false, "Region cannot be updated while running")
				return
			}

			nm.mgm.UpdateRegion(nc.Region)

			nc.SR(true, "Region updated")

		}()

	case "SetEstate":
		//region-estate relationships can only be modified when the region is halted
		go func() {
			stat, ok := nm.mgm.GetRegionStat(nc.Region.UUID)
			if !ok {
				nc.SR(false, "Region not found")
				return
			}
			if stat.Running {
				nc.SR(false, "Region cannot be updated while running")
				return
			}

			nm.mgm.MoveRegionToEstate(nc.Region, nc.Estate)

			nc.SR(true, "Region updated")

		}()

	case "AssignToHost":
		//add a region to a host
		//make sure region is unassigned, and not running
		nm.log.Info("Adding region %v to host %v", nc.Region.UUID.String(), nc.Host.ID)

		r, ok := nm.mgm.GetRegion(nc.Region.UUID)
		if !ok {
			nc.SR(false, "Error: the region was not found")
			continue
		}
		if r.Host != 0 {
			nc.SR(false, "Error, the region is already assigned to a host")
			continue
		}

		st, ok := nm.mgm.GetRegionStat(r.UUID)
		if ok && st.Running {
			nc.SR(false, "Error: The region is running")
			continue
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
			continue
		}
		r, ok := nm.mgm.GetRegion(nc.Region.UUID)
		if !ok {
			nc.SR(false, "Error: region not found")
			continue
		}
		if r.Host != nc.Host.ID {
			nc.SR(false, "Error: The region is not on that host")
			continue
		}

		stat, ok := nm.mgm.GetRegionStat(r.UUID)
		if ok && stat.Running {
			nc.SR(false, "Error: The region is running")
			nm.log.Info("Removing region %v from host %v failed.  The region is currently running", nc.Region.UUID.String(), nc.Host.ID)
			continue
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
		present := false
		for _, h := range conns {
			if h.host.Address == nc.Host.Address {
				nc.SR(false, "Error, Host address already present")
				present = true
			}
		}
		if present {
			continue
		}

		h := nm.mgm.AddHost(nc.Host)
		if h.ID == 0 {
			nc.SR(false, "Error, Host address already present")
			continue
		}
		s := hostSession{
			host: h,
			log:  logger.Wrap(strconv.FormatInt(h.ID, 10), nm.log),
		}
		conns[h.ID] = s

		nc.SR(true, "Host Added")

	default:
		nc.SR(false, "Not Implemented")
	}
*/
//case msg := <-nm.internalMsgs:
/*		switch msg.request {
					case "GetHosts":
						for _, c := range conns {
							msg.hosts <- c.host
						}
						close(msg.hosts)
					}

		}
	}
}
*/

// NodeManager receives and communicates with mgm Node processes
func (m Manager) listen(newConns chan<- hostSession) {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(m.listenPort))
	if err != nil {
		m.log.Fatal("MGM Node listener cannot start: ", err)
		return
	}
	m.listener = ln
	m.log.Info("Listening for mgm host instances on :%d", m.listenPort)

	for {
		conn, err := m.listener.Accept()
		if err != nil {
			m.log.Error("Error accepting connection: ", err)
			continue
		}
		//validate connection, and identify host
		addr := conn.RemoteAddr()
		address := addr.(*net.TCPAddr).IP.String()
		hosts := m.mgm.QueryHosts()
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
			m.log.Error(errmsg)
			conn.Close()
			continue
		}
		if host.Address != address {
			m.log.Info("mgmNode connection from unregistered address: ", address)
			continue
		}
		m.log.Info("MGM Node connection from: %v (%v)", host.ID, address)

		//lookup current assignment of regions
		var regs []mgm.Region
		for _, r := range m.rMgr.GetRegions() {
			if r.Host == host.ID {
				regs = append(regs, r)
			}
		}

		s := hostSession{host: host, conn: conn, regions: regs}
		newConns <- s
	}
}
