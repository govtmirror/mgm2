package webClient

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/m-o-s-e-s/mgm/core"
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/satori/go.uuid"
)

type clientResponse struct {
	MessageID   int
	MessageType string
	Message     interface{}
}

// WebSocketConnector listens for authenticated websocket connections
type WebSocketConnector interface {
	WebsocketHandler(w http.ResponseWriter, r *http.Request)
}

type wsConn struct {
	httpConnector HTTPConnector
	session       chan<- core.UserSession
	logger        logger.Log
}

// NewWebsocketConnector constructs a websocket handler for use
func NewWebsocketConnector(hc HTTPConnector, s chan<- core.UserSession, log logger.Log) WebSocketConnector {
	return wsConn{hc, s, logger.Wrap("WEBSOCK", log)}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (wc wsConn) WebsocketHandler(w http.ResponseWriter, r *http.Request) {

	// test if session exists
	s, _ := wc.httpConnector.GetStore().Get(r, "MGM")
	if len(s.Values) == 0 {
		wc.logger.Info("Websocket closed, no existing session")

		response := clientResponse{MessageType: "AccessDenied", Message: []byte("No Session Found")}
		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}
	// test origin, etc for websocket security
	// not sure if necessary, we will be over https, and the session is valid

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		wc.logger.Error("Error upgrading websocket: ", err)
		return
	}

	guid := s.Values["guid"].(uuid.UUID)
	uLevel := s.Values["ulevel"].(uint8)

	c := client{
		ws,
		make(chan []byte, 64),
		make(chan []byte, 64),
		guid,
		uLevel,
		wc.logger,
		make(chan core.UserObject, 64),
		make(chan bool, 0),
	}
	go c.reader()
	go c.writer()
	go c.processSend()
	wc.session <- c
}

func (c *client) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		c.fromClient <- message
	}
	close(c.fromClient)
	c.ws.Close()
}

func (c *client) writer() {
	for message := range c.toClient {

		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	close(c.toClient)
	c.ws.Close()
}
