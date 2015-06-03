package remote

import (
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Region is a management interface for region processes
type Region interface {
	WriteRegionINI(mgm.Region) error
	WriteOpensimINI([]mgm.ConfigOption, []mgm.ConfigOption) error
	Start()
}

type regionCmd struct {
	command string
	success string
}

type region struct {
	region    mgm.Region
	cmds      chan regionCmd
	log       logger.Log
	dir       string
	hostName  string
	isRunning bool
}

// NewRegion constructs a Region for use
func NewRegion(regionRecord mgm.Region, path string, hostname string, log logger.Log) Region {
	reg := region{}
	reg.region = regionRecord
	reg.cmds = make(chan regionCmd, 8)
	reg.log = logger.Wrap("Region", log)
	reg.dir = path
	reg.hostName = hostname

	go reg.communicate()
	go reg.process()

	return reg
}

func (r region) communicate() {
	for {
		select {
		case cmd := <-r.cmds:
			switch cmd.command {
			case "start":
				r.log.Info("start region goes here")
				//if already running, exit
				if r.isRunning {
					r.log.Error("Region is already running", r.region.UUID)
				}
				//load ini files
				//err := writeOpensimINI(cmd.DefaultConfig, r.dir)
				//execute binaries
			default:
				r.log.Info("Received unexpected command: %v", cmd.command)
			}
		}
	}
}

func (r region) process() {

}

func (r region) Start() {
	cmd := regionCmd{command: "start"}
	r.cmds <- cmd
}
