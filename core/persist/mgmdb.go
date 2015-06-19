package persist

import (
	"fmt"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/simian"
	"github.com/satori/go.uuid"
)

// Notifier is an object that MGMDB will call as data is modified
// Notifier is responsible for alerting interested parties
type Notifier interface {
	HostUpdated(mgm.Host)
	HostDeleted(mgm.Host)
	HostStat(mgm.HostStat)
	RegionUpdated(mgm.Region)
	RegionDeleted(mgm.Region)
	RegionStat(mgm.RegionStat)
}

//MGMDB interfaces with mysql, caches values for performance, and notifies subscribers of object updates
// This interface handles persistance with the mysql database, but also caches non-persisted values as
// a single point of access
type MGMDB interface {
	//host functions
	GetHosts() []mgm.Host
	GetHostStats() []mgm.HostStat
	UpdateHost(mgm.Host)
	UpdateHostStat(mgm.HostStat)
	RemoveHost(mgm.Host)
	//region functions
	GetRegions() []mgm.Region
	GetRegionStats() []mgm.RegionStat
	UpdateRegion(mgm.Region)
	UpdateRegionStat(mgm.RegionStat)
	RemoveRegion(mgm.Region)
	//job functions
	GetJobs() []mgm.Job
	UpdateJob(mgm.Job)
	RemoveJob(mgm.Job)
	//user functions
	GetUsers() []mgm.User
	UpdateUser(mgm.User)
	SetPassword(mgm.User, string)
	GetPendingUsers() []mgm.PendingUser
	//Estate functions
	GetEstates() []mgm.Estate
	//Gropu functions
	GetGroups() []mgm.Group
}

// NewMGMDB constructs an MGMDB instance for use
func NewMGMDB(db Database, osdb Database, sim simian.Connector, log logger.Log, not Notifier) MGMDB {
	mgm := mgmDB{
		db:     db,
		osdb:   osdb,
		sim:    sim,
		log:    logger.Wrap("MGMDB", log),
		reqs:   make(chan mgmReq, 64),
		notify: not,
	}

	go mgm.process()

	return mgm
}

type mgmReq struct {
	request string
	object  interface{}
	result  chan interface{}
}

type mgmDB struct {
	db     Database
	osdb   Database
	sim    simian.Connector
	log    logger.Log
	notify Notifier
	reqs   chan mgmReq
}

func (m mgmDB) process() {
	//populate regions
	regions := make(map[uuid.UUID]mgm.Region)
	for _, r := range m.queryRegions() {
		regions[r.UUID] = r
	}
	//populate hosts
	hosts := make(map[int]mgm.Host)
	hostStats := make(map[int]mgm.HostStat)
	for _, h := range m.queryHosts() {
		hosts[h.ID] = h
		hostStats[h.ID] = mgm.HostStat{}
	}
	//populate users
	users := make(map[uuid.UUID]mgm.User)
	simUsers, err := m.sim.GetUsers()
	if err != nil {
		errMsg := fmt.Sprintf("Cannot read users from simian: %v", err.Error())
		m.log.Fatal(errMsg)
	}
	for _, simUser := range simUsers {
		users[simUser.UserID] = simUser
	}
	simUsers = nil
	pendingUsers := make(map[string]mgm.PendingUser)
	for _, u := range m.queryPendingUsers() {
		pendingUsers[u.Email] = u
	}
	//populate groups
	groups := make(map[uuid.UUID]mgm.Group)
	simGroups, err := m.sim.GetGroups()
	if err != nil {
		errMsg := fmt.Sprintf("Cannot read groups from simian: %v", err.Error())
		m.log.Fatal(errMsg)
	}
	for _, simGroup := range simGroups {
		groups[simGroup.ID] = simGroup
	}
	simGroups = nil
	//populate jobs
	jobs := make(map[int]mgm.Job)
	for _, j := range m.queryJobs() {
		jobs[j.ID] = j
	}
	//populate estates
	estates := make(map[int]mgm.Estate)
	for _, e := range m.queryEstates() {
		estates[e.ID] = e
	}

	for {
		select {
		case req := <-m.reqs:
			switch req.request {
			case "GetRegions":
				for _, r := range regions {
					req.result <- r
				}
				close(req.result)
			case "GetHosts":
				for _, h := range hosts {
					req.result <- h
				}
				close(req.result)
			case "GetHostStats":
				for _, h := range hostStats {
					req.result <- h
				}
				close(req.result)
			case "GetUsers":
				for _, u := range users {
					req.result <- u
				}
				close(req.result)
			case "GetPendingUsers":
				for _, u := range pendingUsers {
					req.result <- u
				}
				close(req.result)
			case "GetJobs":
				for _, j := range jobs {
					req.result <- j
				}
				close(req.result)
			case "GetEstates":
				for _, e := range estates {
					req.result <- e
				}
				close(req.result)
			case "GetGroups":
				for _, g := range groups {
					req.result <- g
				}
				close(req.result)
			case "UpdateHost":
				host := req.object.(mgm.Host)
				hosts[host.ID] = host
				go m.persistHost(host)
				m.notify.HostUpdated(host)
			case "UpdateHostStat":
				stat := req.object.(mgm.HostStat)
				hostStats[stat.ID] = stat
				m.notify.HostStat(stat)
			case "RemoveHost":
				m.log.Info("Removing host: %v", req.object.(mgm.Host).ID)
				host := req.object.(mgm.Host)
				delete(hosts, host.ID)
				go m.purgeHost(host)
				m.notify.HostDeleted(host)
			default:
				errMsg := fmt.Sprintf("Unexpected command: %v", req.request)
				m.log.Error(errMsg)
			}
		}
	}
}
