// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgeproto

import (
	"fmt"
	"github.com/edgexr/edge-cloud/log"
)

func (s *StatusInfo) ToString() string {
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

func (s *StatusInfo) StatusReset() {
	s.TaskNumber = 0
	s.MaxTasks = 0
	s.TaskName = ""
	s.StepName = ""
	s.MsgCount = 0
	s.Msgs = []string{}
}

func (s *StatusInfo) SetMaxTasks(maxtask uint32) {
	s.MaxTasks = maxtask
}

func (s *StatusInfo) SetTask(task string) {
	if s == nil {
		log.WarnLog("nil StatusInfo")
		return
	}
	if task == "" {
		return
	}
	if s.TaskName == task {
		log.DebugLog(log.DebugLevelApi, "StatusInfo task not changed")
		return
	}
	s.TaskNumber++
	s.TaskName = task
	s.StepName = ""
	s.MsgCount++
	s.Msgs = append(s.Msgs, s.ToString())
}

func (s *StatusInfo) SetStep(step string) {
	if s == nil {
		log.WarnLog("nil StatusInfo")
		return
	}
	s.StepName = step
	s.MsgCount++
	s.Msgs = append(s.Msgs, s.ToString())
}
