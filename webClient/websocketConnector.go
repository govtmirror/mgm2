package webClient

import (
  "net/http"
  "github.com/gorilla/websocket"
  "github.com/satori/go.uuid"
  "github.com/M-O-S-E-S/mgm2/core"
  "encoding/json"
)

type clientResponse struct {
  MessageID int
  MessageType string
  Message interface{}
}

type WebsocketConnector struct {
  httpConnector *HttpConnector
  session chan<- core.UserSession
  logger Logger
}

func NewWebsocketConnector(hc *HttpConnector, session chan<- core.UserSession, logger Logger) (*WebsocketConnector) {
  return &WebsocketConnector{hc, session, logger}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (wc WebsocketConnector) WebsocketHandler(w http.ResponseWriter, r *http.Request) {

  // test if session exists
  session, _ := wc.httpConnector.store.Get(r, "MGM")
  if len(session.Values) == 0 {
    wc.logger.Info("Websocket closed, no existing session")

    response := clientResponse{ MessageType: "AccessDenied", Message: "No Session Found"}
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
    wc.logger.Error("Error upgrading websocket: %v", err)
    return
  }

  guid := session.Values["guid"].(uuid.UUID)
  uLevel := session.Values["ulevel"].(uint8)

  c := client{ws, make(chan []byte, 64), make(chan []byte, 64), guid, uLevel, wc.logger}
  go c.reader()
  go c.writer()
  wc.session <- c
}

func (c *client) reader() {
  for {
    _, message, err := c.ws.ReadMessage()
    if err != nil {
      break
    }
    c.fromClient<-message
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
