package mgm

import (
  "fmt"
  "github.com/gorilla/websocket"
  "encoding/json"
)

type Client struct {
  ws *websocket.Conn
  send chan []byte
}

func (c *Client) process() {
  //spin up reader and writer goroutines
  go c.writer()
  go c.reader()
}

func (c *Client) reader() {
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

func (c *Client) writer() {
  for message := range c.send {
    err := c.ws.WriteMessage(websocket.TextMessage, message)
    if err != nil {
      break
    }
  }
  c.ws.Close()
}

type Request struct{
  MessageType string
  Message json.RawMessage
}