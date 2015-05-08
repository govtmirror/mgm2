package core

import "github.com/satori/go.uuid"

// FileUpload is a tuple for sending uploaded files with job id and uploader information
type FileUpload struct {
	JobID int
	User  uuid.UUID
	File  []byte
}
