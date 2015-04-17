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
  // send user information first so client can map uuids to users
  users, err := userConn.GetUsers()
  if err != nil {
    logger.Error("Error lookin up activeuser account: ", err)
  }
  for _, user := range users {
    if user.Suspended && session.GetAccessLevel() < 250 {
      continue
    }
    session.SendUser(user)
  }
  users = nil

  pendingUsers, err := dataStore.GetPendingUsers()
  if err != nil {
    logger.Error("Error lookin up pending user account: ", err)
  }
  for _, user := range pendingUsers {
    session.SendPendingUser(user)
  }
  pendingUsers = nil

  //send regions this user may control
  regions, err := dataStore.GetRegions()
  if err != nil {
    logger.Error("Error lookin up user regions: ", err)
  }
  for _, r := range regions {
    session.SendRegion(r)
  }
  regions = nil

  //send Estate, Group, and Host dataManager
  estates, err := dataStore.GetEstates()
  if err != nil {
    logger.Error("Error lookin up estates: ", err)
  }
  for _, e := range estates {
    session.SendEstate(e)
  }
  estates = nil
  groups, err := userConn.GetGroups()
  if err != nil {
    logger.Error("Error lookin up groups: ", err)
  }
  for _, g := range groups {
    session.SendGroup(g)
  }
  groups = nil
  //only administrative users need host access
  if session.GetAccessLevel() > 249 {
    hosts, err := dataStore.GetHosts()
    if err != nil {
      logger.Error("Error lookin up hosts: ", err)
    }
    for _, h := range hosts {
      session.SendHost(h)
    }
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
        user, err := userConn.GetUserByID(session.GetGuid())
        if err != nil {
          logger.Error("Error lookin up user account: ", err)
        }
        session.SendUser(*user)
      case "GetRegions":
        var regions []Region
        if session.GetAccessLevel() > 250 {
          regions, err = dataStore.GetRegions()
        } else {
          regions, err = dataStore.GetRegionsForUser(session.GetGuid())
        }
        if err != nil {
          logger.Error("Error lookin up user regions: ", err)
        }
        for _, r := range regions {
          session.SendRegion(r)
        }
      case "GetUsers":

      default:
      logger.Error("Error on message from client: ", m.MessageType)
    }
  }
}
