package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Host is an MGM server record
type Host struct {
	ID              int64
	Address         string
	ExternalAddress string
	Hostname        string
	Regions         []uuid.UUID
	Slots           int
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

// HostDeleted is an MGM server record
type HostDeleted struct {
	ID int64
}

// Serialize implements UserObject interface Serialize function
func (h HostDeleted) Serialize() []byte {
	data, _ := json.Marshal(h)
	return data
}

// ObjectType implements UserObject
func (h HostDeleted) ObjectType() string {
	return "HostDeleted"
}

// HostStat holds mgm host statistical info
type HostStat struct {
	ID         int64
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
