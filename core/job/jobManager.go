package job

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Manager manages jobs, updating database, and notifying subscribed parties
type Manager interface {
	FileUploaded(int, uuid.UUID, []byte)
	GetJobByID(int64) (mgm.Job, bool)
	DeleteJob(mgm.Job, core.ServiceRequest)
	CreateLoadIarJob(mgm.User, string) int64
}

type fileUpload struct {
	JobID int64
	User  uuid.UUID
	File  []byte
}

type regionCommand struct {
}

func (jm jobMgr) newRegionCommand() regionCommand {
	rc := regionCommand{}

	return rc
}

// NewManager constructs a jobManager for use
func NewManager(filePath string, hubRegion uuid.UUID, pers persist.MGMDB, log logger.Log) Manager {

	j := jobMgr{}
	j.fileUp = make(chan fileUpload, 32)
	j.localPath = filePath
	j.log = logger.Wrap("JOB", log)
	j.mgm = pers
	j.hub = hubRegion

	j.regionWorkers = make(map[uuid.UUID]chan regionCommand, 8)

	go j.process()

	return j
}

type jobMgr struct {
	fileUp      chan fileUpload
	subscribe   chan chan<- mgm.Job
	unsubscribe chan chan<- mgm.Job
	mgm         persist.MGMDB
	hub         uuid.UUID

	log logger.Log

	regionWorkers map[uuid.UUID]chan regionCommand

	localPath string
}

func (jm jobMgr) FileUploaded(id int, user uuid.UUID, data []byte) {
	jm.log.Info("Received file upload for job %v", id)
	jm.fileUp <- fileUpload{int64(id), user, data}
}

func (jm jobMgr) GetJobByID(id int64) (mgm.Job, bool) {
	jobs := jm.mgm.GetJobs()
	for _, j := range jobs {
		if j.ID == id {
			return j, true
		}
	}
	return mgm.Job{}, false
}

func (jm jobMgr) DeleteJob(j mgm.Job, sr core.ServiceRequest) {
	//perform any file level maintenance, etc...
	for _, job := range jm.mgm.GetJobs() {
		if job.ID == j.ID {
			j = job
		}
	}

	type file struct {
		FileName string
	}

	var f file
	json.Unmarshal([]byte(j.Data), &f)
	if f.FileName != "" {
		//delete files from disk
		err := os.Remove(f.FileName)
		if err != nil {
			jm.log.Error(fmt.Sprintf("Error deleting file %v from job %v: %v", f.FileName, j.ID, err.Error()))
		}
	}

	jm.mgm.RemoveJob(j)
	sr(true, "Job deleted")
}

func (jm jobMgr) GetJobsForUser(user mgm.User) []mgm.Job {
	jobs := jm.mgm.GetJobs()
	var userJobs []mgm.Job
	for _, j := range jobs {
		if j.User == user.UserID {
			userJobs = append(userJobs, j)
		}
	}
	return userJobs
}

func (jm jobMgr) CreateLoadIarJob(owner mgm.User, inventoryPath string) int64 {
	j := mgm.Job{}
	j.Type = "load_iar"
	j.Timestamp = time.Now()
	j.User = owner.UserID

	jd := LoadIarJob{}
	jd.InventoryPath = inventoryPath
	jd.Status = "Created"

	encDat, _ := json.Marshal(jd)
	j.Data = string(encDat)

	return jm.mgm.AddJob(j)
}

func (jm jobMgr) process() {

	for {
		select {

		case s := <-jm.fileUp:
			jm.log.Info("Processing File upload for job %v", s.JobID)
			// look up job
			j, found := jm.GetJobByID(s.JobID)
			if !found {
				//anything could have happened, but the job doesn't seem to exist, drop file
				jm.log.Error(fmt.Sprintf("Error on job file upload, job %v does not exist", s.JobID))
				continue
			}

			//make sure uploader owns the job in question
			if s.User != j.User {
				jm.log.Info("Attempted upload to job %v by %v, owned by %v", j.ID, s.User, j.User)
				continue
			}

			switch j.Type {
			case "load_iar":
				jm.log.Info("Job %v is of type load_iar", s.JobID)
				iarJob := LoadIarJob{}
				err := json.Unmarshal([]byte(j.Data), &iarJob)
				if err != nil {
					jm.log.Info("Error parsing Load Iar job: %v", err.Error())
					continue
				}

				iarJob.Filename = path.Join(jm.localPath, uuid.NewV4().String())
				iarJob.InventoryPath = "/"

				err = ioutil.WriteFile(iarJob.Filename, s.File, 0644)
				if err != nil {
					jm.log.Error("Error writing file: ", err)
					iarJob.Status = "Error writing file"
					data, _ := json.Marshal(iarJob)
					j.Data = string(data)
					jm.mgm.UpdateJob(j)
					continue
				}

				ch, ok := jm.regionWorkers[jm.hub]
				if !ok {
					jm.log.Error("No worker for region found")
					iarJob.Status = "No worker present"
					data, _ := json.Marshal(iarJob)
					j.Data = string(data)
					jm.mgm.UpdateJob(j)
					continue
				}

				iarJob.Status = "In process"
				data, _ := json.Marshal(iarJob)
				j.Data = string(data)
				jm.mgm.UpdateJob(j)

				//dispatch task
				go jm.loadIarTask(j, iarJob, ch)
			default:
				jm.log.Error(fmt.Sprintf("Invalid upload for type %v", j.Type))
			}
		}
	}

}
