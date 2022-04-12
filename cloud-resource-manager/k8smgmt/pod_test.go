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

package k8smgmt

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var podInsuffCpuMessage string = "0/2 nodes are available: 1 Insufficient cpu, 1 Insufficient memory, 1 node(s) had taint {node-role.kubernetes.io/master: }, that the pod didn't tolerate"

var podStates = []string{
	"gh-someapp-758594bfc9-px6bm   Init   <none>   <none>",
	"autoprov-75df8b4cfb-bndnq   Running   <none>   <none>",
	"demo-deployment-78bd547db7-sbqqp   ContainerCreating   <none>   <none>",
	"demo-deployment-7f8445bc4d-8jh2k   Pending   Unschedulable   " + podInsuffCpuMessage,
}

func TestPodStateRegex(t *testing.T) {
	r := regexp.MustCompile(podStateRegString)
	match0 := r.FindStringSubmatch(podStates[0])
	require.NotNil(t, match0)
	require.Equal(t, "Init", match0[2])

	match1 := r.FindStringSubmatch(podStates[1])
	require.NotNil(t, match1)
	require.Equal(t, "Running", match1[2])

	match2 := r.FindStringSubmatch(podStates[2])
	require.NotNil(t, match2)
	require.Equal(t, "<none>", match2[3])
	require.Equal(t, "<none>", match2[4])

	match3 := r.FindStringSubmatch(podStates[3])
	require.NotNil(t, match3)
	require.Equal(t, "Unschedulable", match3[3])
	require.Equal(t, podInsuffCpuMessage, match3[4])
}
