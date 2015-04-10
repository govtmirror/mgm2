package core

import (
  "github.com/satori/go.uuid"
)

type RegionManager struct {
  NewRegions chan Region
  NewClient chan *Client
}

func newRegionManager() RegionManager{
  return RegionManager{
    make(chan Region, 16),
    make(chan *Client, 256),
  }
}

func (rm * RegionManager) newRegion() (*Region){
  r := &Region{}
  r.frames = make(chan int, 64)
  return r
}

func (rm * RegionManager) process(){
  regions := map[uuid.UUID]Region{}
  clients := map[uuid.UUID]*Client{}
  for {
    select{
    case r := <- rm.NewRegions:
      regions[r.UUID] = r
    case c := <- rm.NewClient:
      clients[c.guid] = c
      for k := range regions {
        notifyUserNewRegion(c,regions[k])
      }
    }
  }
}

func notifyUserNewRegion(c *Client, r Region){
  /*type clientRegion struct {
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
  c.send <- data*/
}