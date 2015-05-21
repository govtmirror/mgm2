package mgm

import (
	"encoding/json"

	"github.com/satori/go.uuid"
)

// Region is an opensim region record
type Region struct {
	UUID            uuid.UUID
	Name            string
	Size            uint
	HTTPPort        int       `json:"-"`
	ConsolePort     int       `json:"-"`
	ConsoleUname    uuid.UUID `json:"-"`
	ConsolePass     uuid.UUID `json:"-"`
	LocX            uint
	LocY            uint
	ExternalAddress string `json:"-"`
	SlaveAddress    string `json:"-"`
	IsRunning       bool
	EstateName      string

	frames chan int `json:"-"`
}

func (r *Region) countFrames() {
	/*vals := []int{}
	  start := time.Now()
	  val, more := <- r.frames
	  if !more {
	    r.logger.Info("Region frame counter %v aborting, channel closed\n", r.Name)
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
	    r.logger.Info("Region %v: %.2f fps\n", r.Name, fps)
	  }*/
}

// Serialize implements UserObject interface Serialize function
func (r Region) Serialize() []byte {
	data, _ := json.Marshal(r)
	return data
}

// ObjectType implements UserObject
func (r Region) ObjectType() string {
	return "Region"
}
