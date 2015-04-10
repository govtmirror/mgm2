package core

type Simian interface {
}

type Database interface {
  TestConnection() error
  GetAllRegions() error
}

type Opensim interface {
  Listen()
}

type mgmRequest struct {
  request string
  listener chan<- []byte
}

type mgmCore struct{
  requests chan mgmRequest
  clientMgr ClientManager
}

func NewMGM(simian Simian, database Database, opensim Opensim) (*mgmCore, error){
    
  regionMgr := newRegionManager()
  go regionMgr.process()
    
  clientMgr := ClientManager{}
    
  //opensim := openSimListener{config.OpensimPort, regionMgr}
    
  mgmInstance := &mgmCore{make(chan mgmRequest, 256), clientMgr}
  
  err := database.TestConnection()
  if err != nil {
    return nil, err
  }
    
  err = database.GetAllRegions()
  if err != nil {
    return nil, err
  }
    
  //allow opensim connections
  //go opensim.Listen()
    
  return mgmInstance, nil
}
