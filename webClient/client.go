package webClient

import (
	"encoding/json"

	"github.com/M-O-S-E-S/mgm/core"
	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

type client struct {
	ws           *websocket.Conn
	toClient     chan []byte
	fromClient   chan []byte
	guid         uuid.UUID
	userLevel    uint8
	logger       core.Logger
	toClientChan chan core.UserObject
	closed       bool
}

func (c client) GetSend() chan<- core.UserObject {
	return c.toClientChan
}

func (c client) processSend() {
	for c.closed == false {
		select {
		case msg := <-c.toClientChan:
			c.send(msg)
		}
	}
}

func (c client) send(ob core.UserObject) {
	resp := clientResponse{MessageType: ob.ObjectType(), Message: ob.Serialize()}
	data, err := json.Marshal(resp)
	if err == nil {
		c.writeData(data)
	}
}

func (c client) SignalSuccess(req int, message string) {
	resp := clientResponse{req, "Success", message}
	data, err := json.Marshal(resp)
	if err == nil {
		c.writeData(data)
	}
}

func (c client) SignalError(req int, message string) {
	resp := clientResponse{req, "Error", message}
	data, err := json.Marshal(resp)
	if err == nil {
		c.writeData(data)
	}
}

func (c client) writeData(data []byte) {
	defer func() {
		if x := recover(); x != nil {
			c.logger.Info("Attempt to write to closed client channel")
		}
	}()
	c.toClient <- data
}

func (c client) GetGUID() uuid.UUID {
	return c.guid
}

func (c client) GetAccessLevel() uint8 {
	return c.userLevel
}

func (c client) Read(upstream chan<- []byte, closing chan<- bool) {
	for {
		data, more := <-c.fromClient
		if !more {
			closing <- true
			return
		}
		upstream <- data
	}
}
