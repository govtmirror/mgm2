package job

import (
	"encoding/json"

	"github.com/m-o-s-e-s/mgm/mgm"
)

// LoadIarJob is the data field for jobs that are of type load_iar
type LoadIarJob struct {
	InventoryPath string
	Filename      string
	Status        string
}

//loadIarTask is a coroutine that manages and reports on loading an iar file
func (jm jobMgr) loadIarTask(j mgm.Job, iarJob LoadIarJob, ch chan<- regionCommand) {

	//locate hub region
	var r mgm.Region
	found := false
	for _, reg := range jm.mgm.GetRegions() {
		if reg.UUID == jm.hub {
			found = true
			r = reg
		}
	}
	if !found {
		jm.log.Error("Hub region not found for job")
		iarJob.Status = "Hub region not found"
		data, _ := json.Marshal(iarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}

	//make sure region is running
	for _, stat := range jm.mgm.GetRegionStats() {
		if stat.UUID == r.UUID {
			if !stat.Running {
				jm.log.Error("Hub region not running")
				iarJob.Status = "Hub region not running"
				data, _ := json.Marshal(iarJob)
				j.Data = string(data)
				jm.mgm.UpdateJob(j)
				return
			}
		}
	}

	//locate host for region
	var h mgm.Host
	found = false
	for _, host := range jm.mgm.GetHosts() {
		if host.ID == r.Host {
			found = true
			h = host
		}
	}
	if !found {
		jm.log.Error("Host not found for running hub region...")
		iarJob.Status = "Host not found"
		data, _ := json.Marshal(iarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}

	//we are pretty sure that the host is running, but lets be certain
	for _, stat := range jm.mgm.GetHostStats() {
		if stat.ID == h.ID {
			if !stat.Running {
				jm.log.Error("Hub region not running")
				iarJob.Status = "Hub region not running"
				data, _ := json.Marshal(iarJob)
				j.Data = string(data)
				jm.mgm.UpdateJob(j)
				return
			}
		}
	}

	//insert admin credential on user in question

	//notify worker
	ch <- regionCommand{}

	//close console

	//remove loaded iar file

	//update job

	jm.log.Info("Load iar task exit")
}
