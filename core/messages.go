package core

import "github.com/satori/go.uuid"

// FileUpload is a tuple for sending uploaded files with job id and uploader information
type FileUpload struct {
	JobID int
	User  uuid.UUID
	File  []byte
}

// NetworkMessage is a wrapper for sending multiple message types accross a single wire
type NetworkMessage struct {
	MessageType string
	HStats      HostStats `json:",omitempty"`
}

// HostStats holds mgm host statistical info
type HostStats struct {
	ID         uint
	CPUPercent []float64
	MEMTotal   uint64
	MEMUsed    uint64
	MEMPercent float64
	NetSent    uint64
	NetRecv    uint64
}
