package mgm

import (
  "fmt"
  "github.com/gorilla/websocket"
  "encoding/json"
  "github.com/satori/go.uuid"
)

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