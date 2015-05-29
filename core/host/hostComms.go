package host

import (
	"encoding/json"
	"net"

	"github.com/m-o-s-e-s/mgm/core"
)

// NodeConns is a structure of shared read/write functions between MGM and MGMNode
type HostComms struct {
	Connection net.Conn
	Closing    chan bool
	Log        core.Logger
}

// ReadConnection is a processing loop for reading a socket and parsing messages
func (node HostComms) ReadConnection(readMsgs chan<- core.NetworkMessage) {
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
func (node HostComms) WriteConnection(writeMsgs <-chan core.NetworkMessage) {

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
