package testservices

import (
	"testing"

	"github.com/mobiledgex/edge-cloud/notify"
)

func TestDummySendOrder(t *testing.T) {
	// check send order of dummy handlers
	dh := notify.NewDummyHandler()
	serverMgr := notify.ServerMgr{}
	dh.RegisterServer(&serverMgr)
	CheckNotifySendOrder(t, serverMgr.GetSendOrder())
}
