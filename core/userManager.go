package core

func UserManager(sessionListener <-chan UserSession, dataStore Database, userConn UserConnector, logger Logger){
  
  //create notification hub
  
  
  //listen for user sessions and hook them in
  go func(){
    for {
      select {
        case s := <-sessionListener:
          go userSession(s, dataStore, userConn, logger)
        
      }
    }
  }()
  
}

func userSession(session UserSession, dataStore Database, userConn UserConnector, logger Logger){
  //perform client initialization
  // send initial account information
  accountData, err := userConn.GetUserByID(session.GetGuid())
  if err != nil {
    logger.Error("Error lookin up user account: ", err)
  }
  session.SendUserAccount(*accountData)

  //send regions this user may control
  var regions []Region
  if session.GetAccessLevel() > 250 {
    regions, err = dataStore.GetAllRegions()
  } else {
    regions, err = dataStore.GetRegionsFor(session.GetGuid())
  }
  if err != nil {
    logger.Error("Error lookin up user regions: ", err)
  }
  for _, r := range regions {
    session.SendUserRegion(r)
  }

  for {
    msg, more := session.Read()
    if !more {
      logger.Info("Client went away")
      return
    }
    
    m := userRequest{}
    m.load(msg)
    switch m.MessageType {
      case "GetAccount":
        accountData, err = userConn.GetUserByID(session.GetGuid())
        if err != nil {
          logger.Error("Error lookin up user account: ", err)
        }
        session.SendUserAccount(*accountData)
      case "GetRegions":
        regions, err := dataStore.GetRegionsFor(session.GetGuid())
        if err != nil {
          logger.Error("Error lookin up user account: ", err)
        }
      for _, r := range regions {
        session.SendUserRegion(r)
      }
      case "GetUsers":

      default:
      logger.Error("Error on message from client: ", m.MessageType)
    }
  }
}
