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
	Region      mgm.Region
	MessageType string
	Message     string              `json:",omitempty"`
	Host        mgm.Host            `json:"-"`
	SR          core.ServiceRequest `json:"-"`
}

// ReadConnection is a processing loop for reading a socket and parsing messages
func (node Comms) ReadConnection(readMsgs chan<- core.NetworkMessage) {
	d := json.NewDecoder(node.Connection)

	for {
		nmsg := core.NetworkMessage{}
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
func (node Comms) WriteConnection(writeMsgs <-chan core.NetworkMessage) {

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
