package k8smgmt

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var podStates = []string{
	"gh-someapp-758594bfc9-px6bm   0/1   Init:0/1   0     2s",
	"autoprov-75df8b4cfb-bndnq              1/1       Running   0          30d",
	"demo-deployment-78bd547db7-sbqqp          0/1     ContainerCreating   0          4s",
	"demo-deployment-78bd547db7-sbqqp          0/1     Terminating   0          24s",
}

func TestPodStateRegex(t *testing.T) {
	r := regexp.MustCompile(podStateRegString)
	match0 := r.FindStringSubmatch(podStates[0])
	require.NotNil(t, match0)
	require.Equal(t, "Init:0/1", match0[2])

	match1 := r.FindStringSubmatch(podStates[1])
	require.NotNil(t, match1)
	require.Equal(t, "Running", match1[2])

	match2 := r.FindStringSubmatch(podStates[2])
	require.NotNil(t, match2)
	require.Equal(t, "ContainerCreating", match2[2])

	match3 := r.FindStringSubmatch(podStates[3])
	require.NotNil(t, match3)
	require.Equal(t, "Terminating", match3[2])
}
