package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Host is an MGM server record
type Host struct {
	ID              int
	Address         string
	ExternalAddress string
	Hostname        string
	Regions         []uuid.UUID
	Slots           uint
	Running         bool
}

// Serialize implements UserObject interface Serialize function
func (h Host) Serialize() []byte {
	data, _ := json.Marshal(h)
	return data
}

// ObjectType implements UserObject
func (h Host) ObjectType() string {
	return "Host"
}

// HostStat holds mgm host statistical info
type HostStat struct {
	ID         int
	CPUPercent []float64
	MEMTotal   uint64
	MEMUsed    uint64
	MEMPercent float64
	NetSent    uint64
	NetRecv    uint64
}

// Serialize implements UserObject interface Serialize function
func (h HostStat) Serialize() []byte {
	data, _ := json.Marshal(h)
	return data
}

// ObjectType implements UserObject
func (h HostStat) ObjectType() string {
	return "HostStat"
}
