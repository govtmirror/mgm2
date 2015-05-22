package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// ConfigOption is an opensim.ini configuration line record
type ConfigOption struct {
	Region  uuid.UUID
	Section string
	Item    string
	Content string
}

// Serialize implements UserObject interface Serialize function
func (c ConfigOption) Serialize() []byte {
	data, _ := json.Marshal(c)
	return data
}

// ObjectType implements UserObject
func (c ConfigOption) ObjectType() string {
	return "Config"
}
