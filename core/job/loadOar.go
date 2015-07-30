package job

import (
	"encoding/json"
	"time"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// loadOarJob is the data field for jobs that are of type load_oar
type loadOarJob struct {
	Region   uuid.UUID
	Filename string
	Status   string
}

func (jm jobMgr) CreateLoadOarJob(owner mgm.User, r mgm.Region) int64 {
	j := mgm.Job{}
	j.Type = "load_oar"
	j.Timestamp = time.Now()
	j.User = owner.UserID

	jd := loadOarJob{}
	jd.Region = r.UUID
	jd.Status = "Created"

	encDat, _ := json.Marshal(jd)
	j.Data = string(encDat)

	return jm.mgm.AddJob(j)
}

//loadIarTask is a coroutine that manages and reports on loading an iar file
func (jm jobMgr) loadOarTask(j mgm.Job, oarJob loadOarJob, ch chan<- regionCommand) {

	//locate hub region
	r, found := jm.mgm.GetRegion(oarJob.Region)
	if !found {
		jm.log.Error("Hub region not found for job")
		oarJob.Status = "Hub region not found"
		data, _ := json.Marshal(oarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}

	//make sure region is running
	for _, stat := range jm.mgm.GetRegionStats() {
		if stat.UUID == r.UUID {
			if !stat.Running {
				jm.log.Error("Region not running")
				oarJob.Status = "Region not running"
				data, _ := json.Marshal(oarJob)
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
		jm.log.Error("Host not found for region...")
		oarJob.Status = "Host not found"
		data, _ := json.Marshal(oarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}

	//we are pretty sure that the host is running, but lets be certain
	for _, stat := range jm.mgm.GetHostStats() {
		if stat.ID == h.ID {
			if !stat.Running {
				jm.log.Error("host not running")
				oarJob.Status = "host not running"
				data, _ := json.Marshal(oarJob)
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
