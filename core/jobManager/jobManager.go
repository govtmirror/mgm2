package jobManager

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// JobManager manages jobs, updating database, and notifying subscribed parties
type JobManager interface {
	Subscribe() core.Subscription
	FileUploaded(int, uuid.UUID, []byte)
}

type fileUpload struct {
	JobID int
	User  uuid.UUID
	File  []byte
}

// NewJobManager constructs a jobManager for use
func NewJobManager(filePath string, db core.Database, logger core.Logger) JobManager {

	subscribeChan := make(chan chan<- mgm.Job, 32)
	unsubscribeChan := make(chan chan<- mgm.Job, 32)
	notifyChan := make(chan mgm.Job, 32)

	j := jobMgr{}
	j.fileUp = make(chan fileUpload, 32)
	j.localPath = filePath
	j.log = logger
	j.subscribe = subscribeChan
	j.unsubscribe = unsubscribeChan
	j.broadcast = notifyChan
	j.datastore = db

	j.subs = core.NewSubscriptionManager()

	go j.process()

	return j
}

type jobMgr struct {
	fileUp      chan fileUpload
	subscribe   chan chan<- mgm.Job
	unsubscribe chan chan<- mgm.Job
	broadcast   chan mgm.Job
	datastore   core.Database

	subs core.SubscriptionManager

	log core.Logger

	localPath string
}

func (jm jobMgr) FileUploaded(id int, user uuid.UUID, data []byte) {
	jm.fileUp <- fileUpload{id, user, data}
}

func (jm jobMgr) Subscribe() core.Subscription {
	return jm.subs.Subscribe()
}

func (jm jobMgr) process() {

	go func() {
		for {
			select {

			case s := <-jm.fileUp:
				jm.log.Info("File Upload Received for task %v", s.JobID)
				// look up job
				job, err := jm.datastore.GetJobByID(s.JobID)
				if err != nil {
					//anything could have happened, but the job doesn't seem to exist, drop file
					continue
				}

				//make sure uploader owns the job in question
				if s.User != job.User {
					jm.log.Info("Attempted upload to job %v by %v, owned by %v", job.ID, s.User, job.User)
					continue
				}

				jm.log.Info("Retrieved job %v for %v", job.Type, job.User)

				switch job.Type {
				case "load_iar":
					iarJob := LoadIarJob{}
					err := json.Unmarshal([]byte(job.Data), &iarJob)
					if err != nil {
						jm.log.Info("Error parsing Load Iar job: %v", err.Error())
					}

					iarJob.Filename = uuid.NewV4()

					err = ioutil.WriteFile(path.Join(jm.localPath, iarJob.Filename.String()), s.File, 0644)
					if err != nil {
						jm.log.Error("Error writing file: ", err)
					}

					iarJob.Status = "Iar upload to MGM"

					data, _ := json.Marshal(iarJob)
					job.Data = string(data)

					jm.datastore.UpdateJob(job)

					jm.broadcast <- job
				}
			}
		}
	}()

}
