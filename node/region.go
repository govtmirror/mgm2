package node

import (
	"github.com/M-O-S-E-S/mgm/core"
	"github.com/M-O-S-E-S/mgm/mgm"
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

	log core.Logger
}

// NewRegion constructs a Region for use
func NewRegion(regionRecord mgm.Region, logger core.Logger) Region {
	reg := region{}
	reg.region = regionRecord
	reg.cmds = make(chan regionCmd, 8)

	reg.log = logger

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
