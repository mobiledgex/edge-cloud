package edgeproto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAlertKey(t *testing.T) {
	m := map[string]string{
		"cloudlet":  "localtest",
		"cluster":   "AppCluster",
		"dev":       "MobiledgeX",
		"alertname": "DeadMansSwitch",
		"operator":  "mexdev",
		"severity":  "none",
	}
	key := MapKey(m)
	expectedKey := `{"alertname":"DeadMansSwitch","cloudlet":"localtest","cluster":"AppCluster","dev":"MobiledgeX","operator":"mexdev","severity":"none"}`
	require.Equal(t, expectedKey, key)

	a := Alert{}
	AlertKeyStringParse(key, &a)
	require.Equal(t, m, a.Labels)
}
