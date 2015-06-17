package mgm

import (
	"encoding/json"
	"time"

	"github.com/satori/go.uuid"
)

// Region is an opensim region record
type Region struct {
	UUID         uuid.UUID
	Name         string
	Size         uint
	HTTPPort     int
	ConsolePort  int
	ConsoleUname uuid.UUID
	ConsolePass  uuid.UUID
	LocX         uint
	LocY         uint
	Host         int
	IsRunning    bool
	EstateName   string

	frames chan int
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
	type clientSafeRegion struct {
		UUID       uuid.UUID
		Name       string
		Size       uint
		LocX       uint
		LocY       uint
		Host       int
		IsRunning  bool
		EstateName string
	}
	csr := clientSafeRegion{r.UUID, r.Name, r.Size, r.LocX, r.LocY, r.Host, r.IsRunning, r.EstateName}
	data, _ := json.Marshal(csr)
	return data
}

// ObjectType implements UserObject
func (r Region) ObjectType() string {
	return "Region"
}

// RegionStat holds region-specific runtime metrics
type RegionStat struct {
	UUID       uuid.UUID
	Running    bool
	CPUPercent float64
	MemKB      float64
	Uptime     time.Duration
}

// Serialize implements UserObject interface Serialize function
func (rs RegionStat) Serialize() []byte {
	data, _ := json.Marshal(rs)
	return data
}

// ObjectType implements UserObject
func (rs RegionStat) ObjectType() string {
	return "RegionStat"
}
