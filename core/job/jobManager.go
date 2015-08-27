package job

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

type fileUpload struct {
	JobID int64
	User  uuid.UUID
	File  []byte
}

func (jm Manager) newRegionCommand() regionCommand {
	rc := regionCommand{}

	return rc
}

type notifier interface {
}

// NewManager constructs a jobManager for use
func NewManager(filePath string, mgmURL string, hubRegion uuid.UUID, pers persist.MGMDB, notify notifier, log logger.Log) Manager {

	j := Manager{}
	j.fileUp = make(chan fileUpload, 32)
	j.localPath = filePath
	j.mgmURL = mgmURL
	j.log = logger.Wrap("JOB", log)
	j.mgm = pers
	j.hub = hubRegion
	j.rUp = make(chan uuid.UUID, 32)
	j.rDn = make(chan uuid.UUID, 32)
	j.notify = notify

	j.jobs = make(map[int64]mgm.Job)
	for _, t := range pers.QueryJobs() {
		j.jobs[t.ID] = t
	}
	j.jMutex = &sync.Mutex{}

	go j.process()

	return j
}

// Manager is a central access point for Job operations
type Manager struct {
	fileUp      chan fileUpload
	subscribe   chan chan<- mgm.Job
	unsubscribe chan chan<- mgm.Job
	mgm         persist.MGMDB
	hub         uuid.UUID
	jobs        map[int64]mgm.Job
	jMutex      *sync.Mutex
	notify      notifier

	rUp chan uuid.UUID
	rDn chan uuid.UUID

	log logger.Log

	localPath string
	mgmURL    string
}

// FileUploaded is a handler for uploaded files from clients or nodes
func (jm Manager) FileUploaded(id int, user uuid.UUID, data []byte) {
	jm.log.Info("Received file upload for job %v", id)
	jm.fileUp <- fileUpload{int64(id), user, data}
}

// GetJobByID retrieves a job record matching a specific id
func (jm Manager) GetJobByID(id int64) (mgm.Job, bool) {
	jm.jMutex.Lock()
	defer jm.jMutex.Unlock()
	t, ok := jm.jobs[id]
	return t, ok
}

// DeleteJob purges a job from the cache and database
func (jm Manager) DeleteJob(j mgm.Job) {
	jm.jMutex.Lock()
	defer jm.jMutex.Unlock()
	j, ok := jm.jobs[j.ID]
	if !ok {
		return
	}
	delete(jm.jobs, j.ID)
	jm.mgm.PurgeJob(j)

	//perform any file level maintenance, etc...
	type file struct {
		File string
	}

	var f file
	json.Unmarshal([]byte(j.Data), &f)
	if f.File != "" {
		//delete files from disk
		err := os.Remove(f.File)
		if err != nil {
			jm.log.Error(fmt.Sprintf("Error deleting file %v from job %v: %v", f.File, j.ID, err.Error()))
		}
	}
}

// GetJobsForUser retrieves all jobs owned by a specified user
func (jm Manager) GetJobsForUser(user uuid.UUID) []mgm.Job {
	jm.jMutex.Lock()
	defer jm.jMutex.Unlock()
	userJobs := []mgm.Job{}
	for _, j := range jm.jobs {
		if j.User == user {
			userJobs = append(userJobs, j)
		}
	}
	return userJobs
}

func (jm Manager) process() {

	regionWorkers := make(map[uuid.UUID]chan regionCommand, 8)

	for {
		select {

		case id := <-jm.rUp:
			_, ok := regionWorkers[id]
			if !ok {
				regionWorkers[id] = make(chan regionCommand, 32)
				go jm.processWorker(id, regionWorkers[id])
			}
		case id := <-jm.rDn:
			ch, ok := regionWorkers[id]
			if ok {
				close(ch)
				delete(regionWorkers, id)
			}
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
			/*	jm.log.Info("Job %v is of type load_iar", s.JobID)
				iarJob := loadIarJob{}
				err := json.Unmarshal([]byte(j.Data), &iarJob)
				if err != nil {
					jm.log.Info("Error parsing Load Iar job: %v", err.Error())
					continue
				}

				if iarJob.File != "" {
					jm.log.Info("Job %v multiple upload rejected", err.Error())
					continue
				}

				iarJob.File = path.Join(jm.localPath, uuid.NewV4().String())

				err = ioutil.WriteFile(iarJob.File, s.File, 0644)
				if err != nil {
					jm.log.Error("Error writing file: ", err)
					iarJob.Status = "Error writing file"
					data, _ := json.Marshal(iarJob)
					j.Data = string(data)
					jm.mgm.UpdateJob(j)
					continue
				}

				ch, ok := regionWorkers[jm.hub]
				if !ok {
					jm.log.Error("No worker for region found")
					iarJob.Status = "No worker found: Hub region not running"
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
			*/
			case "load_oar":
			/*	jm.log.Info("Job %v is of type load_oar", s.JobID)
				oarJob := loadOarJob{}
				err := json.Unmarshal([]byte(j.Data), &oarJob)
				if err != nil {
					jm.log.Info("Error parsing Load Oar job: %v", err.Error())
					continue
				}

				if oarJob.File != "" {
					jm.log.Info("Job %v multiple upload rejected", err.Error())
					continue
				}

				oarJob.File = path.Join(jm.localPath, uuid.NewV4().String())

				err = ioutil.WriteFile(oarJob.File, s.File, 0644)
				if err != nil {
					jm.log.Error("Error writing file: ", err)
					oarJob.Status = "Error writing file"
					data, _ := json.Marshal(oarJob)
					j.Data = string(data)
					jm.mgm.UpdateJob(j)
					continue
				}

				ch, ok := regionWorkers[oarJob.Region]
				if !ok {
					jm.log.Error(fmt.Sprintf("No worker for region %v found", oarJob.Region))
					oarJob.Status = "No worker found: Region not running"
					data, _ := json.Marshal(oarJob)
					j.Data = string(data)
					jm.mgm.UpdateJob(j)
					continue
				}

				oarJob.Status = "In process"
				data, _ := json.Marshal(oarJob)
				j.Data = string(data)
				jm.mgm.UpdateJob(j)

				//dispatch task
				go jm.loadOarTask(j, oarJob, ch)
			*/
			default:
				jm.log.Error(fmt.Sprintf("Invalid upload for type %v", j.Type))
			}
		}
	}

}
