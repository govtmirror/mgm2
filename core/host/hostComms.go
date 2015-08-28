package host

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Comms is a structure of shared read/write functions between MGM and MGMNode
type Comms struct {
	Connection *websocket.Conn
	Closing    chan bool
	Log        logger.Log
}

// Message is a messagestructure for MGM<->node messages
type Message struct {
	ID          uint
	MessageType string
	response    chan<- error
	Region      mgm.Region         `json:",omitempty"`
	Message     string             `json:",omitempty"`
	Register    Registration       `json:",omitempty"`
	HStats      mgm.HostStat       `json:",omitempty"`
	RStats      mgm.RegionStat     `json:",omitempty"`
	Configs     []mgm.ConfigOption `json:",omitempty"`
	Host        mgm.Host           `json:"-"`
	Estate      mgm.Estate         `json:"-"`
}

// Registration holds mgmNode information for MGM
type Registration struct {
	ExternalAddress string
	Name            string
	Slots           int
}

var wsupgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// WShandler is a websocket entry point for host connections
func (m Manager) WShandler(w http.ResponseWriter, r *http.Request) {
	conn, err := wsupgrader.Upgrade(w, r, nil)
	if err != nil {
		m.log.Info("Failed to set websocket upgrade: %+v", err)
		return
	}

	go process(&m, conn)

	for {
		t, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		conn.WriteMessage(t, msg)
	}
}

func process(m *Manager, c *websocket.Conn) {
	log := logger.Wrap("comms", m.log)
	//signal channel
	ch := make(chan bool)
	in := make(chan Message, 32)
	//read on the socket
	go func() {
		for {
			msg := Message{}
			err := c.ReadJSON(&msg)
			if err != nil {
				log.Error(err.Error())
				return
			}
			in <- msg
		}
	}()
	//do not write on the socket, that happens elsewhere

	//process packets
	for {
		select {
		case <-ch:
			log.Info("Processing loop exiting")
			return
		case msg := <-in:
			fmt.Println(msg)
			c.WriteJSON(msg)
		}
	}
}
