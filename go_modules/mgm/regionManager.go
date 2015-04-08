package mgm

import (
  "github.com/satori/go.uuid"
  "encoding/json"
  "fmt"
)

type regionManager struct {
  newRegions chan region
  newClient chan *client
}

func newRegionManager() regionManager{
  return regionManager{
    make(chan region, 16),
    make(chan *client, 256),
  }
}

func (rm * regionManager) newRegion() (*region){
  r := &region{}
  r.frames = make(chan int, 64)
  return r
}

func (rm * regionManager) process(){
  regions := map[uuid.UUID]region{}
  clients := map[uuid.UUID]*client{}
  for {
    select{
    case r := <- rm.newRegions:
      regions[r.uuid] = r
    case c := <- rm.newClient:
      clients[c.guid] = c
      for k := range regions {
        notifyUserNewRegion(c,regions[k])
      }
    }
  }
}

func notifyUserNewRegion(c *client, r region){
  type clientRegion struct {
    UUID string
    Name string
    Size uint
    LocX uint
    LocY uint
    IsRunning bool
  }
  
  cr := &clientRegion{
    r.uuid.String(),
    r.name,
    r.size,
    r.locX,
    r.locY,
    r.isRunning,
  }
  
  cm := &clientResponse{
    "NewRegion",
    cr,
  }
    
  data, err := json.Marshal(cm)
  if err != nil {
    fmt.Println(err)
    return
  }
  c.send <- data
}