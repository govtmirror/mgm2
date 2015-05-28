package job

import "github.com/satori/go.uuid"

// LoadIarJob is the data field for jobs that are of type load_iar
type LoadIarJob struct {
	InventoryPath string
	Filename      uuid.UUID
	Status        string
}
