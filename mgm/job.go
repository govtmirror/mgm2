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

// Serialize implements UserObject interface Serialize function
func (j Job) Serialize() []byte {
	data, _ := json.Marshal(j)
	return data
}

// ObjectType implements UserObject
func (j Job) ObjectType() string {
	return "Job"
}
