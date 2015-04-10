package core

import (
  //"github.com/satori/go.uuid"
  "fmt"
  "time"
)

type RegionManager struct {

  database Database
}

func (rm *RegionManager) Process(){
  
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