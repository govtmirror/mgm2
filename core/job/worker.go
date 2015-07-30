package job

import (
	"fmt"

	"github.com/m-o-s-e-s/mgm/core/logger"
	"github.com/satori/go.uuid"
)

func (jm jobMgr) processWorker(id uuid.UUID, cmds <-chan regionCommand) {
	log := logger.Wrap(id.String(), jm.log)

	log.Info("Begin Processing")
	for {
		select {
		case cmd, ok := <-cmds:
			if !ok {
				//command channel closed, exit processing
				log.Info("Channel closed, exiting")
				return
			}
			jm.log.Info(fmt.Sprintf("Received Command %v", cmd))
		}
	}
}
