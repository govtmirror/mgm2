package core

import (
  "github.com/satori/go.uuid"
  "fmt"
)

type UserSession struct {
  ToClient chan interface{}
  FromClient chan []byte
  Guid uuid.UUID
}

type UserSource interface {
  GetUserByID(uuid.UUID) (User, error)
}

func UserManager(sessionListener <-chan UserSession, dataStore Database, userSource UserSource){
  
  //create notification hub
  
  
  //listen for user sessions and hook them in
  go func(){
    for {
      select {
        case s := <-sessionListener:
          go userSession(s, dataStore, userSource)
        
      }
    }
  }()
  
}

func userSession(session UserSession, dataStore Database, userSource UserSource){
  //perform client initialization
  //request account information
  accountData, err := userSource.GetUserByID(session.Guid)
  if err != nil {
    fmt.Println("Error lookin up user account: ", err)
  }
  session.ToClient <- accountData
  fmt.Println(accountData)
  //lookup what this user can control



  for {
    msg, more := <-session.FromClient
    if !more {
      fmt.Println("Client went away")
      return
    }
    
    m := userRequest{}
    m.load(msg)
    switch m.MessageType {
      case "GetAccount":

      case "GetRegions":

      case "GetUsers":

      default:
      fmt.Println("Error on message from client: ", m.MessageType)
    }
  }
}
