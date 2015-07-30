package job

import (
	"fmt"
	"strings"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/m-o-s-e-s/mgm/core/region"
	"github.com/satori/go.uuid"
)

type regionCommand struct {
	command string
	filter  string
	success string
	failure string
	respond chan<- response
}

type response struct {
	success bool
	message string
}

func (jm jobMgr) processWorker(id uuid.UUID, cmds <-chan regionCommand) {
	log := logger.Wrap(id.String(), jm.log)

	log.Info("Begin Processing")
	for {
		cmd, ok := <-cmds
		if !ok {
			//command channel closed, exit processing
			log.Info("Channel closed, exiting")
			return
		}

		//locate region and host objects
		r, ok := jm.mgm.GetRegion(id)
		if !ok {
			cmd.respond <- response{false, "Region no longer exists"}
			continue
		}
		h, ok := jm.mgm.GetHost(r.Host)
		if !ok {
			cmd.respond <- response{false, "Region is not on a host"}
			continue
		}

		//make sure they are both still running
		rStat, ok := jm.mgm.GetRegionStat(r.UUID)
		if !ok || !rStat.Running {
			cmd.respond <- response{false, "Region is not running"}
			continue
		}
		hStat, ok := jm.mgm.GetHostStat(h.ID)
		if !ok || !hStat.Running {
			cmd.respond <- response{false, "Host is not running"}
			continue
		}

		//connect to the console
		c, err := region.NewRestConsole(r, h)
		if err != nil {
			cmd.respond <- response{false, fmt.Sprintf("Could not connecto to console: %v", err.Error())}
			continue
		}
		c.Write(cmd.command)
		succeeded := false
		message := "Disconnected"
		for {
			msg, ok := <-c.Read()
			if !ok {
				break
			}
			log.Info(msg)
			if !strings.Contains(msg, cmd.filter) {
				continue
			}
			if strings.Contains(msg, cmd.success) {
				succeeded = true
				message = "Completed"
			}
			if strings.Contains(msg, cmd.failure) {
				message = "Failed"
			}
		}
		c.Close()

		cmd.respond <- response{succeeded, message}

		jm.log.Info(fmt.Sprintf("Received Command %v", cmd))

	}
}
