package persist

import (
	"fmt"
	"time"

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
	EstateUpdated(mgm.Estate)
	EstateDeleted(mgm.Estate)
	JobUpdated(mgm.Job)
	JobDeleted(j mgm.Job)
}

//MGMDB interfaces with mysql, caches values for performance, and notifies subscribers of object updates
// This interface handles persistance with the mysql database, but also caches non-persisted values as
// a single point of access
type MGMDB interface {
	//host functions
	GetHosts() []mgm.Host
	GetHostStats() []mgm.HostStat
	AddHost(mgm.Host) mgm.Host
	UpdateHost(mgm.Host)
	UpdateHostStat(mgm.HostStat)
	RemoveHost(mgm.Host)
	//region functions
	GetRegions() []mgm.Region
	GetRegionStats() []mgm.RegionStat
	UpdateRegion(mgm.Region)
	UpdateRegionStat(mgm.RegionStat)
	RemoveRegion(mgm.Region)
	MoveRegionToEstate(mgm.Region, mgm.Estate)
	MoveRegionToHost(mgm.Region, mgm.Host)
	GetDefaultConfigs() []mgm.ConfigOption
	GetConfigs(mgm.Region) []mgm.ConfigOption
	//job functions
	GetJobs() []mgm.Job
	AddJob(j mgm.Job) int64
	UpdateJob(mgm.Job)
	RemoveJob(mgm.Job)
	//user functions
	GetUsers() []mgm.User
	UpdateUser(mgm.User)
	SetPassword(mgm.User, string)
	GetPendingUsers() []mgm.PendingUser
	//Estate functions
	GetEstates() []mgm.Estate
	//Group functions
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
	target  interface{}
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
	regionStats := make(map[uuid.UUID]mgm.RegionStat)
	for _, r := range m.queryRegions() {
		regions[r.UUID] = r
	}
	//populate hosts
	hosts := make(map[int64]mgm.Host)
	hostStats := make(map[int64]mgm.HostStat)
	for _, h := range m.queryHosts() {
		hosts[h.ID] = h
		hostStats[h.ID] = mgm.HostStat{}
	}
	for _, r := range regions {
		h, ok := hosts[r.Host]
		if ok {
			h.Regions = append(h.Regions, r.UUID)
			hosts[h.ID] = h
		}
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
	jobs := make(map[int64]mgm.Job)
	for _, j := range m.queryJobs() {
		jobs[j.ID] = j
	}
	//populate estates
	estates := make(map[int64]mgm.Estate)
	for _, e := range m.queryEstates() {
		estates[e.ID] = e
	}

ProcessingPackets:
	for {
		select {
		case req := <-m.reqs:
			switch req.request {
			case "GetRegions":
				for _, r := range regions {
					req.result <- r
				}
				close(req.result)
			case "GetRegionStats":
				for _, r := range regionStats {
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
			case "AddJob":
				job := req.object.(mgm.Job)
				job.Timestamp = time.Now()
				id, err := m.insertJob(job)
				if err != nil {
					errMsg := fmt.Sprintf("Error adding job: %v", err.Error())
					m.log.Error(errMsg)
					req.result <- mgm.Job{}
					close(req.result)
					continue
				}
				job.ID = id
				jobs[id] = job
				m.notify.JobUpdated(job)
				req.result <- job
				close(req.result)
			case "RemoveJob":
				job := req.object.(mgm.Job)
				_, ok := jobs[job.ID]
				if ok {
					delete(jobs, job.ID)
					go m.purgeJob(job)
					m.notify.JobDeleted(job)
				}
			case "AddHost":
				host := req.object.(mgm.Host)
				//inserts are not asynchronous, as we need the insert ID to populate ourselves
				host.ID, err = m.insertHost(host)
				if err != nil {
					errMsg := fmt.Sprintf("Error adding host: %v", err.Error())
					m.log.Error(errMsg)
					req.result <- mgm.Host{}
					close(req.result)
					continue
				}
				hosts[host.ID] = host
				m.notify.HostUpdated(host)
				req.result <- host
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
			case "UpdateRegion":
				region := req.object.(mgm.Region)
				r, ok := regions[region.UUID]
				if !ok {
					errMsg := fmt.Sprintf("Error updating region: %v, Region does not exist", region.UUID)
					m.log.Error(errMsg)
				}
				if region.Host != r.Host {
					//host has been updated, update relevant host records
					//remove region from host r
					if r.Host != 0 {
						h, ok := hosts[r.Host]
						if ok {
							//remove record
							for i, id := range h.Regions {
								if id == r.UUID {
									h.Regions = append(h.Regions[:i], h.Regions[i+1:]...)
									hosts[h.ID] = h
									m.notify.HostUpdated(h)
								}
							}
						}
					}
					//add region to host region
					if region.Host != 0 {
						h, ok := hosts[region.Host]
						if ok {
							h.Regions = append(h.Regions, region.UUID)
							hosts[h.ID] = h
							m.notify.HostUpdated(h)
						}
					}
				}
				regions[region.UUID] = region
				go m.persistRegion(region)
				m.notify.RegionUpdated(region)
			case "UpdateRegionStat":
				stat := req.object.(mgm.RegionStat)
				regionStats[stat.UUID] = stat
				m.notify.RegionStat(stat)
			case "RemoveHost":
				host := req.object.(mgm.Host)
				delete(hosts, host.ID)
				go m.purgeHost(host)
				m.notify.HostDeleted(host)
			case "MoveRegionToHost":
				reg := req.object.(mgm.Region)
				host := req.target.(mgm.Host)
				h := hosts[host.ID]
				r := regions[reg.UUID]
				//make sure there is room in the new hosts
				if len(h.Regions) >= h.Slots && h.ID != 0 {
					errMsg := fmt.Sprintf("Host %v already has all slots filled", h.ID)
					m.log.Error(errMsg)
					break ProcessingPackets
				}
				if r.Host != 0 {
					//remove region from current host
					host = hosts[r.Host]
					for i, id := range host.Regions {
						if id != r.UUID {
							host.Regions = append(host.Regions[:i], host.Regions[i+1:]...)
							hosts[host.ID] = host
							m.notify.HostUpdated(host)
						}
					}
				}
				//place region on new host
				if h.ID != 0 {
					h.Regions = append(h.Regions, r.UUID)
					hosts[h.ID] = h
					m.notify.HostUpdated(h)
				}

				//update region record
				r.Host = h.ID
				regions[r.UUID] = r
				m.notify.RegionUpdated(r)

				go m.persistRegion(r)

			case "MoveRegionToEstate":
				reg := req.object.(mgm.Region)
				est := req.target.(mgm.Estate)
				//check if region already in estate
				for _, id := range estates[est.ID].Regions {
					if id == reg.UUID {
						req.result <- false
						req.result <- "Region is already in that estate"
						close(req.result)
						break ProcessingPackets
					}
				}
				//persist change
				go m.persistRegionEstate(reg, est)
				//remove region from current estate
				for _, e := range estates {
					for y, id := range e.Regions {
						if id == reg.UUID {
							m.log.Error("Removing region from current estate object")
							e.Regions = append(e.Regions[:y], e.Regions[y+1:]...)
							estates[e.ID] = e
							m.notify.EstateUpdated(e)
						}
					}
				}

				//add region to new estate
				est = estates[est.ID]
				est.Regions = append(est.Regions, reg.UUID)
				estates[est.ID] = est
				m.notify.EstateUpdated(est)
				req.result <- true
				req.result <- "estate updated"
				close(req.result)
			case "GetDefaultConfigs":
				go func() {
					con, err := m.db.GetConnection()
					if err != nil {
						errMsg := fmt.Sprintf("Error loading default configs: %v", err.Error())
						m.log.Error(errMsg)
						close(req.result)
						return
					}
					defer con.Close()

					rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region IS NULL")
					if err != nil {
						errMsg := fmt.Sprintf("Error loading default configs: %v", err.Error())
						m.log.Error(errMsg)
						close(req.result)
						return
					}
					defer rows.Close()

					for rows.Next() {
						c := mgm.ConfigOption{}
						err = rows.Scan(
							&c.Section,
							&c.Item,
							&c.Content,
						)
						if err != nil {
							errMsg := fmt.Sprintf("Error loading default configs: %v", err.Error())
							m.log.Error(errMsg)
							close(req.result)
							return
						}
						req.result <- c
					}
					close(req.result)
				}()
			case "GetConfigs":
				go func() {
					region := req.object.(mgm.Region)
					con, err := m.db.GetConnection()
					if err != nil {
						errMsg := fmt.Sprintf("Error loading default configs: %v", err.Error())
						m.log.Error(errMsg)
						close(req.result)
						return
					}
					defer con.Close()

					rows, err := con.Query("SELECT section, item, content FROM iniConfig WHERE region=?", region.UUID.String())
					if err != nil {
						errMsg := fmt.Sprintf("Error loading default configs: %v", err.Error())
						m.log.Error(errMsg)
						close(req.result)
						return
					}
					defer rows.Close()

					for rows.Next() {
						c := mgm.ConfigOption{}
						err = rows.Scan(
							&c.Section,
							&c.Item,
							&c.Content,
						)
						if err != nil {
							errMsg := fmt.Sprintf("Error loading default configs: %v", err.Error())
							m.log.Error(errMsg)
							close(req.result)
							return
						}
						c.Region = region.UUID
						req.result <- c
					}
					close(req.result)
				}()
			default:
				errMsg := fmt.Sprintf("Unexpected command: %v", req.request)
				m.log.Error(errMsg)
			}
		}
	}
}
