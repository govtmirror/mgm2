package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Group is an Opensim group record
type Group struct {
	Name    string
	Founder uuid.UUID
	ID      uuid.UUID
	Members []uuid.UUID
	Roles   []string
}

// Serialize implements UserObject interface Serialize function
func (g Group) Serialize() []byte {
	data, _ := json.Marshal(g)
	return data
}

// ObjectType implements UserObject
func (g Group) ObjectType() string {
	return "Group"
}
