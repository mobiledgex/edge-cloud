package edgeproto

import (
	"fmt"
	"github.com/mobiledgex/edge-cloud/log"
)

func (s *StatusInfo) toString() string {
	if s == nil {
		return ""
	}
	if s.TaskNumber > 0 {
		rc := s.TaskName
		if s.MaxTasks > 0 {
			rc += fmt.Sprintf("(%d of %d)", s.TaskNumber, s.MaxTasks)
		}
		if s.StepName != "" {
			rc += fmt.Sprintf(", %s", s.StepName)
		}
		return rc
	}
	return ""
}

func (s *StatusInfo) reset() {
	s.TaskNumber = 0
	s.MaxTasks = 0
	s.TaskName = ""
	s.StepName = ""
}

func (s *StatusInfo) setMaxTasks(maxtask uint32) {
	s.MaxTasks = maxtask
}

func (s *StatusInfo) setTask(task string) {
	if s == nil {
		log.WarnLog("nil StatusInfo")
		return
	}
	if s.TaskName == task {
		log.DebugLog(log.DebugLevelApi, "StatusInfo task not changed")
		return
	}
	s.TaskNumber++
	s.TaskName = task
	s.StepName = ""
}

func (s *StatusInfo) setStep(step string) {
	if s == nil {
		log.WarnLog("nil StatusInfo")
		return
	}
	s.StepName = step
}
