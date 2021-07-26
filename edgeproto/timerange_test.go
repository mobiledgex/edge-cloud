package edgeproto

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeRange(t *testing.T) {
	now := time.Now()
	tr := TimeRange{
		StartTime: now,
		EndTime:   now.Add(-time.Second),
	}
	err := tr.Resolve(time.Second)
	require.NotNil(t, err)
	require.Equal(t, "start time must be before (older than) end time", err.Error())

	tr = TimeRange{
		StartAge: Duration(time.Second),
		EndAge:   Duration(2 * time.Second),
	}
	err = tr.Resolve(time.Second)
	require.NotNil(t, err)
	require.Equal(t, "start age must be before (older than) end age", err.Error())
}
