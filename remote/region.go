package remote

import (
	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/mgm"
)

// Region is a management interface for region processes
type Region interface {
}

type regionCmd struct {
	ID      int
	command string
	success string
}

type region struct {
	region mgm.Region

	cmds chan regionCmd

	log logger.Log
}

// NewRegion constructs a Region for use
func NewRegion(regionRecord mgm.Region, log logger.Log) Region {
	reg := region{}
	reg.region = regionRecord
	reg.cmds = make(chan regionCmd, 8)

	reg.log = logger.Wrap("Region", log)

	go reg.communicate()
	go reg.process()
	return reg
}

func (r region) communicate() {
	for {
		select {
		case cmd := <-r.cmds:
			r.log.Info("Received command: %v", cmd.command)
		}
	}
}

func (r region) process() {

}
