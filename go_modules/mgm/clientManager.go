package mgm

import (
  "fmt"
  "github.com/gorilla/websocket"
  "github.com/satori/go.uuid"
  "time"
)

type ClientManager struct {
  authIn chan clientAuth
  authTest chan clientAuthRequest
}

type clientAuth struct {
  guid uuid.UUID 
  token uuid.UUID
  address string
}

type clientAuthRequest struct {
  client clientAuth
  callback chan <- bool
}

func (r *ClientManager) NewClient(ws *websocket.Conn) *Client {
  fmt.Println("New client constructed")
  client := &Client{send: make(chan []byte, 256), ws: ws}
  return client
}

func (r ClientManager) addAuthenticatedUser(client clientAuth){
  r.authIn <- client
}

func (r ClientManager) isUserAuthenticated(client clientAuth) bool{
  response := make(chan bool)
  r.authTest <- clientAuthRequest{client, response }
  return <- response
}

func (r ClientManager) Listen() {
  type authRecord struct {
    client clientAuth
    time time.Time
  }
  users := map[uuid.UUID]authRecord{}
  ticker := time.NewTicker(60 * time.Second)
  //read from two channels, one to add authenticators, one to test authenticators
  //read from recurring timer, to expire authentication requests
  for {
    select {
    case msg:= <- r.authIn:
      //new authentication for user
      users[msg.guid] = authRecord{msg, time.Now()}
    case msg:= <- r.authTest:
      //test user against current authentications
      record, ok := users[msg.client.guid]
      if !ok {
        msg.callback <- false
      }
      if record.client == msg.client {
        msg.callback <- true
      }
      msg.callback <- false
    case <- ticker.C:
      counter := 0
      for record := range users {
        //purge record if older than 1 hour
        if time.Since(users[record].time).Seconds() > 3600 { 
          delete(users, users[record].client.guid)
          counter++
        }
      }
    }
  }
}