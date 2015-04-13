package webClient

import (
  "fmt"
  "net/http"
  "github.com/gorilla/websocket"
  "github.com/satori/go.uuid"
  "github.com/M-O-S-E-S/mgm2/core"
  "encoding/json"
)

type client struct {
  ws *websocket.Conn
  toClient chan interface{}
  fromClient chan []byte
}

type WebsocketConnector struct {
  httpConnector *HttpConnector
  session chan<- core.UserSession
}

func NewWebsocketConnector(hc *HttpConnector, session chan<- core.UserSession) (*WebsocketConnector) {
  return &WebsocketConnector{hc, session}
}

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (wc WebsocketConnector) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New connection on ws")
  
  session, _ := wc.httpConnector.store.Get(r, "MGM")
  // test if session exists
  if len(session.Values) == 0 {
    fmt.Println("Websocket closed, existing session missing")
    return
  }
  // test origin, etc for websocket security
  // not sure if necessary, we will be over https, and the session is valid

  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  
  guid, _ := uuid.FromString( session.Values["guid"].(string))
  
  c := client{ws, make(chan interface{}, 64), make(chan []byte, 64)}
  go c.reader()
  go c.writer()
  wc.session <- core.UserSession{c.toClient, c.fromClient, guid}
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

    data, err := json.Marshal(message)
    if err != nil {
      fmt.Println("Error encoding message: ", err)
      continue
    }


    err = c.ws.WriteMessage(websocket.TextMessage, data)
    if err != nil {
      break
    }
  }
  close(c.toClient)
  c.ws.Close()
}
