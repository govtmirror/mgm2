package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Estate is an opensim estate record
type Estate struct {
	Name     string
	ID       int
	Owner    uuid.UUID
	Managers []uuid.UUID
	Regions  []uuid.UUID
}

// Serialize implements UserObject interface Serialize function
func (e Estate) Serialize() []byte {
	data, _ := json.Marshal(e)
	return data
}

// ObjectType implements UserObject
func (e Estate) ObjectType() string {
	return "Estate"
}
