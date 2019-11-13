package vault

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDummyServer(t *testing.T) {
	server, config := DummyServer()
	defer server.Close()

	data := make(map[string]interface{})
	err := GetData(config, "some path", 0, &data)
	require.Nil(t, err)
}
