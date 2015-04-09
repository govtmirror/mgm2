package mgm

import (
  "time"
  "fmt"
  "github.com/satori/go.uuid"
)

type region struct {
  uuid uuid.UUID
  name string
  size uint
  httpPort int
  consolePort int
  consoleUname uuid.UUID
  consolePass uuid.UUID
  locX uint
  locY uint
  externalAddress string
  slaveAddress string
  isRunning bool
  status string
  
  frames chan int
}

func (r *region) Register(message map[string]interface{}) {
  r.name = message["name"].(string)
  r.locX = uint(message["locX"].(float64))
  r.locY = uint(message["locY"].(float64))
  r.size = uint(message["size"].(float64))
    
  r.frames = make(chan int, 256)
  go r.countFrames()
    
  fmt.Printf("Region %s registered with size %v at %v, %v.\n", r.name, r.size, r.locX, r.locY)
}

func (r *region) Cleanup() {
  close(r.frames)
}

func (r *region) countFrames() {
  vals := []int{}
  start := time.Now()
  val, more := <- r.frames
  if !more {
    fmt.Printf("Region frame counter %v aborting, channel closed\n", r.name)   
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
    fmt.Printf("Region %v: %.2f fps\n", r.name, fps)
  }
}