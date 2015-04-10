package core

import (
  "github.com/satori/go.uuid"
  "fmt"
)

type UserSession struct {
  ToClient chan []byte
  FromClient chan []byte
  Guid uuid.UUID
}

type Authenticator interface {

}

func UserManager(sessionListener <-chan UserSession, dataStore Database,auth Authenticator){
  
  //create notification hub
  
  
  //listen for user sessions and hook them in
  go func(){
    for {
      select {
        case s := <-sessionListener:
          go userSession(s, dataStore)
        
      }
    }
  }()
  
}

func userSession(session UserSession, dataStore Database){
  for {
    msg, more := <-session.FromClient
    if !more {
      fmt.Println("Client went away")
      return
    }
    
    m := userRequest{}
    m.load(msg)
    switch m.MessageType {
      default:
      fmt.Println("Error on message from client: ", m.MessageType)
    }
  }
}
