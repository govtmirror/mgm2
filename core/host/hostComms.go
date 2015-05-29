package host

import (
	"encoding/json"
	"net"

	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Comms is a structure of shared read/write functions between MGM and MGMNode
type Comms struct {
	Connection net.Conn
	Closing    chan bool
	Log        core.Logger
}

// Message is a messagestructure for MGM<->node messages
type Message struct {
	MessageType string
	Region      mgm.Region          `json:",omitempty"`
	Message     string              `json:",omitempty"`
	Register    Registration        `json:",omitempty"`
	HStats      mgm.HostStat        `json:",omitempty"`
	Host        mgm.Host            `json:"-"`
	SR          core.ServiceRequest `json:"-"`
}

// Registration holds mgmNode information for MGM
type Registration struct {
	ExternalAddress string
	Name            string
	Slots           uint
}

// ReadConnection is a processing loop for reading a socket and parsing messages
func (node Comms) ReadConnection(readMsgs chan<- Message) {
	d := json.NewDecoder(node.Connection)

	for {
		nmsg := Message{}
		err := d.Decode(&nmsg)
		if err != nil {
			if err.Error() == "EOF" {
				close(node.Closing)
				node.Connection.Close()
				return
			}
			node.Log.Error("Error decoding mgm message: ", err)
		}

		readMsgs <- nmsg
	}
}

// WriteConnection is a processing loop for json encoding messages to a socket
func (node Comms) WriteConnection(writeMsgs <-chan Message) {

	for {
		select {
		case <-node.Closing:
			return
		case msg := <-writeMsgs:
			data, _ := json.Marshal(msg)
			_, err := node.Connection.Write(data)
			if err != nil {
				node.Connection.Close()
				return
			}
		}
	}
}
