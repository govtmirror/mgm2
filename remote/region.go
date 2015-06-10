package remote

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
	"github.com/satori/go.uuid"
	"github.com/shirou/gopsutil/process"
)

// Region is a management interface for region processes
type Region interface {
	WriteRegionINI(mgm.Region) error
	WriteOpensimINI([]mgm.ConfigOption) error
	Start()
}

type regionCmd struct {
	command string
	success string
}

type region struct {
	UUID      uuid.UUID
	cmds      chan regionCmd
	log       logger.Log
	dir       string
	hostName  string
	rStat     chan<- mgm.RegionStat
	isRunning bool
}

// NewRegion constructs a Region for use
func NewRegion(rID uuid.UUID, path string, hostname string, rStat chan<- mgm.RegionStat, log logger.Log) Region {
	reg := region{}
	reg.UUID = rID
	reg.cmds = make(chan regionCmd, 8)
	reg.log = logger.Wrap(rID.String(), log)
	reg.dir = path
	reg.rStat = rStat
	reg.hostName = hostname

	go reg.communicate()

	return reg
}

func (r region) communicate() {

	ticker := time.NewTicker(5 * time.Second)
	pidFile := path.Join(r.dir, "moses.pid")

	for {
		select {
		case cmd := <-r.cmds:
			switch cmd.command {
			case "start":
				//if already running, exit
				if r.isRunning {
					r.log.Error("Region is already running", r.UUID)
					continue
				}
				//execute binaries
				os.Chdir(r.dir)
				cmdName := "/usr/bin/mono"
				cmdArgs := []string{"OpenSim.exe"}
				cmd := exec.Command(cmdName, cmdArgs...)
				err := cmd.Start()
				if err != nil {
					r.log.Error("Error starting process: %s", err.Error())
				} else {
					r.log.Info("Started Successfully")
					r.isRunning = true
				}
			default:
				r.log.Info("Received unexpected command: %v", cmd.command)
			}
		case <-ticker.C:
			stat := mgm.RegionStat{UUID: r.UUID}
			r.isRunning = false
			stat.Running = false
			//test for region state
			idBytes, err := ioutil.ReadFile(pidFile)
			if err != nil {
				r.rStat <- stat
				continue
			}
			//pid exists
			pid, err := strconv.ParseInt(strings.TrimSpace(string(idBytes)), 10, 32)
			if err != nil {
				r.log.Error("PID contains non-integer content: %s: %s", string(idBytes), err.Error())
				r.rStat <- stat
				continue
			}
			proc, err := process.NewProcess(int32(pid))
			if err != nil {
				r.log.Error("Error creating psutil process.  May not exist")
				r.rStat <- stat
				continue
			}
			cpuPercent, err := proc.CPUPercent(0)
			if err != nil {
				r.log.Error("Error getting cpu for pid: %s", err.Error())
			} else {
				stat.CPUPercent = cpuPercent
			}
			memInfo, err := proc.MemoryInfo()
			if err != nil {
				r.log.Error("Error getting memory for pid: %s", err.Error())
			} else {
				stat.MemKB = (float64(memInfo.RSS) / 1024.0)
			}
			r.log.Info("Proc MemB: %v", memInfo.RSS)
			r.log.Info("Proc MemKB: %v", float64(memInfo.RSS)/1024.0)

			// having trouble pinning this one down.  It moves, and CreateTime should be the process's created time, which is static
			//ct, err := proc.CreateTime()

			r.rStat <- stat
		}
	}
}

func (r region) Start() {
	cmd := regionCmd{command: "start"}
	r.cmds <- cmd
}
