package edgeproto

import "github.com/mobiledgex/edge-cloud/log"
import "fmt"

func (s *StatusInfo) toString() string {
	if s == nil {
		return ""
	}
	if s.TaskNumber > 0 {
		rc := fmt.Sprintf("TASK: [%02d] - [%s]", s.TaskNumber, s.TaskName)
		if s.StepName != "" {
			rc += fmt.Sprintf("\n   -- STEP: [%s]", s.StepName)
		}
		return rc
	}
	return ""
}

func (s *StatusInfo) reset() {
	log.DebugLog(log.DebugLevelApi, "StatusInfo resetting status")
	s.TaskNumber = 0
	s.TaskName = ""
	s.StepName = ""
}

func (s *StatusInfo) setTask(task string) {
	log.DebugLog(log.DebugLevelApi, "StatusInfo setting task", "task", task)
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
	log.DebugLog(log.DebugLevelApi, "StatusInfo setting step", "step", step)
	if s == nil {
		log.WarnLog("nil StatusInfo")
		return
	}
	s.StepName = step
}
