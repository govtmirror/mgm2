package host

import (
	"errors"
	"net"
	"sync"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/sql"
)

type notifier interface {
	HostRemoved(int64)
	HostUpdated(mgm.Host)
	HostStat(mgm.HostStat)
}

type hostConn struct {
}

func (hc hostConn) Close() {

}

// NewManager constructs NodeManager instances
func NewManager(port int, rMgr region.Manager, pers sql.MGMDB, notify notifier, log logger.Log) Manager {
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
	mgm             sql.MGMDB
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

// UpdateHostStats consume an updated host stat, notifying the client manager as well
func (m Manager) UpdateHostStats(hs mgm.HostStat) {
	m.hsMutex.Lock()
	defer m.hsMutex.Unlock()
	m.hostStats[hs.ID] = hs
	m.notify.HostStat(hs)
}
