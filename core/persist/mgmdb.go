package persist

import (
	"fmt"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/m-o-s-e-s/mgm/simian"
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

// NewMGMDB constructs an MGMDB instance for use
func NewMGMDB(db Database, osdb Database, sim simian.Connector, log logger.Log) MGMDB {
	mgm := MGMDB{
		db:   db,
		osdb: osdb,
		sim:  sim,
		log:  logger.Wrap("MGMDB", log),
		reqs: make(chan mgmReq, 64),
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

// MGMDB is a central acecss point for MGMs internal cache, and also handles persistence asynchronously
type MGMDB struct {
	db   Database
	osdb Database
	sim  simian.Connector
	log  logger.Log
	reqs chan mgmReq
}

func (m MGMDB) process() {

	for {
		select {
		case req := <-m.reqs:
			switch req.request {
			case "MoveRegionToEstate":
				/*reg := req.object.(mgm.Region)
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
				*/
			case "GetConfigs":
				go func() {
					region := req.object.(mgm.Region)
					con, err := m.db.getConnection()
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
