package job

import (
	"time"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/persist"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
)

// Manager manages jobs, updating database, and notifying subscribed parties
type Manager interface {
	FileUploaded(int, uuid.UUID, []byte)
	GetJobByID(int) (mgm.Job, bool)
	DeleteJob(mgm.Job)
	CreateLoadIarJob(mgm.User, string)
	GetJobsForUser(mgm.User) []mgm.Job
}

type fileUpload struct {
	JobID int
	User  uuid.UUID
	File  []byte
}

// NewManager constructs a jobManager for use
func NewManager(filePath string, pers persist.MGMDB, log logger.Log) Manager {

	j := jobMgr{}
	j.fileUp = make(chan fileUpload, 32)
	j.localPath = filePath
	j.log = logger.Wrap("JOB", log)
	j.mgm = pers

	go j.process()

	return j
}

type jobMgr struct {
	fileUp      chan fileUpload
	subscribe   chan chan<- mgm.Job
	unsubscribe chan chan<- mgm.Job
	broadcast   chan mgm.Job
	mgm         persist.MGMDB

	log logger.Log

	localPath string
}

func (jm jobMgr) FileUploaded(id int, user uuid.UUID, data []byte) {
	jm.fileUp <- fileUpload{id, user, data}
}

func (jm jobMgr) GetJobByID(id int) (mgm.Job, bool) {
	jobs := jm.mgm.GetJobs()
	for _, j := range jobs {
		if j.ID == id {
			return j, true
		}
	}
	return mgm.Job{}, false
}

func (jm jobMgr) DeleteJob(j mgm.Job) {
	jm.mgm.RemoveJob(j)
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

func (jm jobMgr) CreateLoadIarJob(owner mgm.User, inventoryPath string) {
	j := mgm.Job{}
	j.Type = "load_iar"
	j.Timestamp = time.Now()
	j.User = owner.UserID
	j.Data = inventoryPath
	jm.mgm.UpdateJob(j)
}

func (jm jobMgr) process() {

	/*go func() {
		for {
			select {

			case s := <-jm.fileUp:
				jm.log.Info("File Upload Received for task %v", s.JobID)
				// look up job
				j, err := jm.db.GetJobByID(s.JobID)
				if err != nil {
					//anything could have happened, but the job doesn't seem to exist, drop file
					continue
				}

				//make sure uploader owns the job in question
				if s.User != j.User {
					jm.log.Info("Attempted upload to job %v by %v, owned by %v", j.ID, s.User, j.User)
					continue
				}

				jm.log.Info("Retrieved job %v for %v", j.Type, j.User)

				switch j.Type {
				case "load_iar":
					iarJob := LoadIarJob{}
					err := json.Unmarshal([]byte(j.Data), &iarJob)
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
					j.Data = string(data)

					jm.db.UpdateJob(j)

					jm.broadcast <- j
				}
			}
		}
	}()*/

}
