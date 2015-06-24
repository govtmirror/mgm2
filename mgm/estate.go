package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Estate is an opensim estate record
type Estate struct {
	Name     string
	ID       int64
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

// EstateDeleted is an opensim estate record
type EstateDeleted struct {
	ID int64
}

// Serialize implements UserObject interface Serialize function
func (e EstateDeleted) Serialize() []byte {
	data, _ := json.Marshal(e)
	return data
}

// ObjectType implements UserObject
func (e EstateDeleted) ObjectType() string {
	return "EstateDeleted"
}
