package webClient

import (
  "fmt"
  "net/http"
  "github.com/gorilla/websocket"
  "encoding/json"
  //"github.com/satori/go.uuid"
)

type WebsocketConnector struct {
  httpConnector *HttpConnector
}

func NewWebsocketConnector(hc *HttpConnector) (*WebsocketConnector) {
  return &WebsocketConnector{hc}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (wc WebsocketConnector) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New connection on ws")
  
  session, _ := wc.httpConnector.store.Get(r, "MGM")
  // test if session exists
  if len(session.Values) == 0 {
    fmt.Println("Websocket closed, existing session missing");
    return
  }
  // test origin, etc for websocket security
  // not sure if necessary, we will be over https, and the session is valid
  
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  
  c := newClient(ws)
  //cm.regionMgr.NewClient <- c
  c.process()
}

func newClient(ws *websocket.Conn) *client{
  return &client{ws, make(chan []byte, 64)}
}

type client struct {
  ws *websocket.Conn
  send chan []byte
}

func (c *client) process() {
  //spin up reader and writer goroutines
  go c.writer()
  go c.reader()
}

func (c *client) reader() {
  for {
    _, message, err := c.ws.ReadMessage()
    if err != nil {
      break
    }
    type userRequest struct {
      MessageType string
      Message json.RawMessage
    }
    var m userRequest
    err = json.Unmarshal(message, &m)
    if err != nil {
      fmt.Println("Error decoding message: ", err)
      continue
    }
    fmt.Println("Message received with type: ", m.MessageType)
    //c.send<-message
  }
  fmt.Println("reader closing connection")
  c.ws.Close()
}

func (c *client) writer() {
  for message := range c.send {
    err := c.ws.WriteMessage(websocket.TextMessage, message)
    if err != nil {
      break
    }
  }
  c.ws.Close()
}

type clientRequest struct{
  MessageType string
  Message json.RawMessage
}

type clientResponse struct {
  MessageType string
  Message interface{}
}