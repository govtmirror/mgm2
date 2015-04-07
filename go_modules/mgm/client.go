package mgm

import (
  "fmt"
  "github.com/gorilla/websocket"
  "net/http"
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
        fmt.Println(message)
        c.send<-message
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

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (cm ClientManager) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New connection on ws")
  
  session, _ := cm.store.Get(r, "MGM")
  if len(session.Values) == 0 {
    fmt.Println("Websocket closed, existing session missing");
    return
  }
  
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  c := cm.NewClient(ws)
  c.process()
}