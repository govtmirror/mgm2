package mgm

import (
    "github.com/gorilla/websocket"
    "fmt"
    "net/http"
)

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type ClientHandler struct {
    ClientMgr ClientManager
}

func (ch ClientHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    fmt.Println("New connection on ws")
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println(err)
        return
    }
    c := ch.ClientMgr.NewClient(ws)
    c.process()
}