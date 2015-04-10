package core

import (
  "time"
  "fmt"
  "github.com/satori/go.uuid"
)

type Region struct {
  UUID uuid.UUID
  Name string
  Size uint
  HttpPort int
  ConsolePort int
  ConsoleUname uuid.UUID
  ConsolePass uuid.UUID
  LocX uint
  LocY uint
  ExternalAddress string
  SlaveAddress string
  IsRunning bool
  Status string
  
  frames chan int
}

func (r *Region) Register(message map[string]interface{}) {
  r.Name = message["name"].(string)
  r.LocX = uint(message["locX"].(float64))
  r.LocY = uint(message["locY"].(float64))
  r.Size = uint(message["size"].(float64))
    
  r.frames = make(chan int, 256)
  go r.countFrames()
    
  fmt.Printf("Region %s registered with size %v at %v, %v.\n", r.Name, r.Size, r.LocX, r.LocY)
}

func (r *Region) Cleanup() {
  close(r.frames)
}

func (r *Region) countFrames() {
  vals := []int{}
  start := time.Now()
  val, more := <- r.frames
  if !more {
    fmt.Printf("Region frame counter %v aborting, channel closed\n", r.Name)   
    return
  }
  vals = append(vals,val)
  for {
    val, more = <- r.frames
    if !more {
      return
    }
    vals = append(vals,val)
    elapsed := time.Since(start).Seconds()
    if elapsed < 5 {
      continue
    }
    start = time.Now()
    first := vals[0]
    last := vals[len(vals)-1]
    fps := float64(len(vals)) / ((float64(last) - float64(first))/1000.0)
    vals = vals[len(vals)-1:]
    fmt.Printf("Region %v: %.2f fps\n", r.Name, fps)
  }
}