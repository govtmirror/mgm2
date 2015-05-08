package core

import (
	"encoding/json"
	"io/ioutil"
	"path"

	"github.com/satori/go.uuid"
)

// JobManager is the Job processing goroutine for MGM
func JobManager(fileUpload <-chan FileUpload, notify chan<- Job, rootPath string, dataStore Database, logger Logger) {

	go func() {
		for {
			select {

			case s := <-fileUpload:
				logger.Info("File Upload Received for task %v", s.JobID)
				// look up job
				job, err := dataStore.GetJobByID(s.JobID)
				if err != nil {
					//anything could have happened, but the job doesn't seem to exist, drop file
					continue
				}

				//make sure uploader owns the job in question
				if s.User != job.User {
					logger.Info("Attempted uplooad to job %v by %v, owned by %v", job.ID, s.User, job.User)
					continue
				}

				logger.Info("Retrieved job %v for %v", job.Type, job.User)

				switch job.Type {
				case "load_iar":
					iarJob := LoadIarJob{}
					err := json.Unmarshal([]byte(job.Data), &iarJob)
					if err != nil {
						logger.Info("Error parsing Load Iar job: %v", err.Error())
					}

					iarJob.Filename = uuid.NewV4()

					err = ioutil.WriteFile(path.Join(rootPath, iarJob.Filename.String()), s.File, 0644)
					if err != nil {
						logger.Error("Error writing file: ", err)
					}

					iarJob.Status = "Iar upload to MGM"

					data, _ := json.Marshal(iarJob)
					job.Data = string(data)

					dataStore.UpdateJob(job)

					notify <- job
				}

			}
		}
	}()

}
