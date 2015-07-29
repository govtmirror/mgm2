package job

import "fmt"

func (jm jobMgr) processWorker(cmds <-chan regionCommand) {
	for {
		select {
		case cmd, ok := <-cmds:
			if !ok {
				//command channel closed, exit processing
				return
			}
			jm.log.Info(fmt.Sprintf("Received Command %v", cmd))
		}
	}
}
