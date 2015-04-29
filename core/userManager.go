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

  for {
    msg, more := session.Read()
    if !more {
      logger.Info("Client went away")
      return
    }
    m := userRequest{}
    m.load(msg)
    switch m.MessageType {
      case "SetPassword":
        userID, password, err := m.readPassword()
        if err != nil {
          logger.Error("Error reading password request")
          continue
        }
        logger.Info("Setting password for %v to %v", userID, password)
        if userID != session.GetGuid() && session.GetAccessLevel() < 250 {
          session.SignalError(m.MessageID, "Permission Denied")
        } else {
          if password == "" {
            session.SignalError(m.MessageID, "Password Cannot be blank")
          } else {
            err = userConn.SetPassword(session.GetGuid(), password)
            if err != nil {
              session.SignalError(m.MessageID, err.Error())
            } else {
              session.SignalSuccess(m.MessageID)
            }
          }
        }
      case "GetDefaultConfig":
        if session.GetAccessLevel() > 249 {
          logger.Info("Serving Default Region Configs.  Request: %v", m.MessageID)
          cfgs, err := dataStore.GetDefaultConfigs()
          if err != nil {
            logger.Error("Error getting default configs: %v", err)
          } else {
            for _, cfg := range cfgs {
              session.SendConfig(m.MessageID, cfg)
            }
            session.SignalSuccess(m.MessageID)
          }
        }
      case "GetConfig":
        if session.GetAccessLevel() > 249 {
          logger.Info("Serving Region Configs.  Request: %v", m.MessageID)
          rid, err := m.readRegionID()
          if(err != nil){
            logger.Error("Error reading region id for configs: %v", err)
          } else {
            logger.Info("Serving Region Configs for %v.", rid)
            cfgs, err := dataStore.GetConfigs(rid)
            if err != nil {
              logger.Error("Error getting configs: %v", err)
            } else {
              for _, cfg := range cfgs {
                session.SendConfig(m.MessageID, cfg)
              }
              session.SignalSuccess(m.MessageID)
            }
          }
        }
      case "GetState":
        logger.Info("Service state request")
        users, err := userConn.GetUsers()
        if err != nil {
          logger.Error("Error lookin up activeuser account: ", err)
        }
        for _, user := range users {
          if user.Suspended && session.GetAccessLevel() < 250 {
            continue
          }
          session.SendUser(0, user)
        }
        users = nil

        pendingUsers, err := dataStore.GetPendingUsers()
        if err != nil {
          logger.Error("Error lookin up pending user account: ", err)
        }
        for _, user := range pendingUsers {
          session.SendPendingUser(0, user)
        }
        pendingUsers = nil

        //send regions this user may control
        regions, err := dataStore.GetRegions()
        if err != nil {
          logger.Error("Error lookin up user regions: ", err)
        }
        for _, r := range regions {
          session.SendRegion(0, r)
        }
        regions = nil

        //send Estate, Group, and Host dataManager
        estates, err := dataStore.GetEstates()
        if err != nil {
          logger.Error("Error lookin up estates: ", err)
        }
        for _, e := range estates {
          session.SendEstate(0, e)
        }
        estates = nil
        groups, err := userConn.GetGroups()
        if err != nil {
          logger.Error("Error lookin up groups: ", err)
        }
        for _, g := range groups {
          session.SendGroup(0, g)
        }
        groups = nil
        //only administrative users need host access
        if session.GetAccessLevel() > 249 {
          hosts, err := dataStore.GetHosts()
          if err != nil {
            logger.Error("Error lookin up hosts: ", err)
          }
          for _, h := range hosts {
            session.SendHost(0, h)
          }
        }

        //signal to the client that we have completed initial state sync
        session.SignalSuccess(m.MessageID)
        logger.Info("Sync Complete")

      default:
        logger.Error("Error on message from client: ", m.MessageType)
    }
  }
}
