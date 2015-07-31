package job

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// loadOarJob is the data field for jobs that are of type load_oar
type loadOarJob struct {
	Region   uuid.UUID
	Filename string
	File     string
	Status   string
	Name     string
	X        uint
	Y        uint
	Merge    bool
}

func (jm jobMgr) CreateLoadOarJob(owner mgm.User, r mgm.Region, x uint, y uint, merge bool, filename string) int64 {
	j := mgm.Job{}
	j.Type = "load_oar"
	j.Timestamp = time.Now()
	j.User = owner.UserID

	jd := loadOarJob{}
	jd.Region = r.UUID
	jd.Status = "Created"
	jd.X = x
	jd.Y = y
	jd.Merge = merge
	jd.Filename = filename

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
	h, ok := jm.mgm.GetHost(r.Host)
	if !ok {
		jm.log.Error("Host not found for region...")
		oarJob.Status = "Host not found"
		data, _ := json.Marshal(oarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}

	//we are pretty sure that the host is running, but lets be certain
	stat, ok := jm.mgm.GetHostStat(h.ID)
	if !ok {
		jm.log.Error(fmt.Sprintf("host %v stats not found", h.ID))
		oarJob.Status = "host stats not found"
		data, _ := json.Marshal(oarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}
	if !stat.Running {
		jm.log.Error(fmt.Sprintf("host %v not running", h.ID))
		oarJob.Status = "host not running"
		data, _ := json.Marshal(oarJob)
		j.Data = string(data)
		jm.mgm.UpdateJob(j)
		return
	}

	//generate command
	url := fmt.Sprintf("http://%v/download/%v", jm.mgmURL, j.ID)
	var cmd string
	if oarJob.Merge {
		cmd = fmt.Sprintf("load oar --merge --force-terrain --force-parcels --displacement <%v,%v,0> %v",
			oarJob.X,
			oarJob.Y,
			url,
		)
	} else {
		cmd = fmt.Sprintf("load oar --force-terrain --force-parcels --displacement <%v,%v,0> %v",
			oarJob.X,
			oarJob.Y,
			url,
		)
	}

	//notify worker
	resp := make(chan response)
	ch <- regionCommand{
		command: cmd,
		filter:  "[ARCHIVER]",
		success: "Successfully",
		failure: "Aborting",
		respond: resp,
	}

	//remove loaded iar file

	//update job

}
