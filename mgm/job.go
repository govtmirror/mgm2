package mgm

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"
)

// Job is a record for long-running user tasks in MGM
type Job struct {
	ID        int64
	Timestamp time.Time
	Type      string
	User      uuid.UUID
	Data      string
}

// JobData is an unfortunate struct for encoding job parts into a single database field
// All job data structs have these fields in common for download operations
type JobData struct {
	Status   string
	Filename string
	File     string
}

// ReadData retrieves the JobData struct form our extra data field
func (j Job) ReadData() JobData {
	jd := JobData{}
	json.Unmarshal([]byte(j.Data), &jd)
	return jd
}

// Serialize implements UserObject interface Serialize function
func (j Job) Serialize() []byte {
	data, _ := json.Marshal(j)
	return data
}

// ObjectType implements UserObject
func (j Job) ObjectType() string {
	return "Job"
}

// JobDeleted is an MGM server record
type JobDeleted struct {
	ID int64
}

// Serialize implements UserObject interface Serialize function
func (h JobDeleted) Serialize() []byte {
	data, _ := json.Marshal(h)
	return data
}

// ObjectType implements UserObject
func (h JobDeleted) ObjectType() string {
	return "JobDeleted"
}
