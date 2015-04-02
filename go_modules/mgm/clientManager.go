package mgm

import (
    "fmt"
    "github.com/gorilla/websocket"
)

type ClientManager struct {}

func (r *ClientManager) NewClient(ws *websocket.Conn) *Client {
    fmt.Println("New client constructed")
    client := &Client{send: make(chan []byte, 256), ws: ws}
    return client
}