package core

import (
  "fmt"
)

func UserManager(sessionListener <-chan UserSession, dataStore Database, userConn UserConnector){
  
  //create notification hub
  
  
  //listen for user sessions and hook them in
  go func(){
    for {
      select {
        case s := <-sessionListener:
          go userSession(s, dataStore, userConn)
        
      }
    }
  }()
  
}

func userSession(session UserSession, dataStore Database, userConn UserConnector){
  //perform client initialization
  // send initial account information
  accountData, err := userConn.GetUserByID(session.GetGuid())
  if err != nil {
    fmt.Println("Error lookin up user account: ", err)
  }
  session.SendUserAccount(accountData)

  //send regions this user may control
  regions, err := dataStore.GetRegionsFor(session.GetGuid())
  if err != nil {
    fmt.Println("Error lookin up user account: ", err)
  }
  for _, r := range regions {
    session.SendUserRegion(r)
  }



  for {
    msg, more := session.Read()
    if !more {
      fmt.Println("Client went away")
      return
    }
    
    m := userRequest{}
    m.load(msg)
    switch m.MessageType {
      case "GetAccount":
        accountData, err = userConn.GetUserByID(session.GetGuid())
        if err != nil {
          fmt.Println("Error lookin up user account: ", err)
        }
        session.SendUserAccount(accountData)
      case "GetRegions":
        regions, err := dataStore.GetRegionsFor(session.GetGuid())
        if err != nil {
          fmt.Println("Error lookin up user account: ", err)
        }
      for _, r := range regions {
        session.SendUserRegion(r)
      }
      case "GetUsers":

      default:
      fmt.Println("Error on message from client: ", m.MessageType)
    }
  }
}
