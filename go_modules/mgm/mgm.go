package mgm

import (
  "../simian"
  "fmt"
)

type mgmRequest struct {
  request string
  listener chan<- []byte
}

type MgmConfig struct {
  SimianUrl string
  SessionSecret string
  OpensimPort string
}

type mgmCore struct{
  requests chan mgmRequest
  ClientMgr ClientManager
}

var mgmInstance *mgmCore = nil

func NewMGM(config MgmConfig) (*mgmCore, error){
  if mgmInstance == nil {
    //Make sure that simian is happy
    fmt.Println("Initializing Simiangrid Connection")
    err := simian.Initialize(config.SimianUrl)
    if err != nil {
      return nil, err
    }
    
    //Instantiate our client manager with session keyMGM
    clientMgr := ClientManager{}
    clientMgr.Initialize(config.SessionSecret)
    
    //start listening for opensim connections
    opensim := OpenSimListener{config.OpensimPort}
    go opensim.Listen()
    
    mgmInstance = &mgmCore{make(chan mgmRequest, 256), clientMgr}
  }
  return mgmInstance, nil
}