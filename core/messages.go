package core

import "github.com/m-o-s-e-s/mgm/mgm"

// NetworkMessage is a wrapper for sending multiple message types accross a single wire
type NetworkMessage struct {
	MessageType string
	HStats      mgm.HostStat `json:",omitempty"`

	Region mgm.Region `json:",omitempty"`
}
