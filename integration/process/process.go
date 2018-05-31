package process

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// ProcessSetup defines a collection of various processes.
// Processes are all the processes that make up our edge cloud, such as
// the etcd db, controller, DME, CRM, plus possibly external
// processes for application clients and servers.
//
// The goal of this abstraction layer is to allow the same integration
// tests to be run against different implementations of the processes.
// I.e. for unit testing, processes may be run locally on a laptop
// in the global namespace. For CI/CD testing, the same tests may be
// run against processes in docker images that will be the same images
// used in deployment. Alternatively tests could run against pre-existing
// processes (deployments) already running in the cloud.
type ProcessSetup struct {
	Etcds       []EtcdProcess
	Controllers []ControllerProcess
	Dmes        []DmeProcess
	Crms        []CrmProcess
}

type EtcdProcess interface {
	Start() error
	Stop()
	ResetData() error
}

type ControllerProcess interface {
	Start(opts ...StartOp) error
	Stop()
	ConnectAPI(timeout time.Duration) (*grpc.ClientConn, error)
}

type DmeProcess interface {
	Start() error
	Stop()
	ConnectAPI() (*grpc.ClientConn, error)
}

type CrmProcess interface {
	Start() error
	Stop()
	ConnectAPI() (*grpc.ClientConn, error)
}

// options

type StartOptions struct {
	Debug string
}

type StartOp func(op *StartOptions)

func WithDebug(debug string) StartOp {
	return func(op *StartOptions) { op.Debug = debug }
}

func (s *StartOptions) ApplyStartOptions(opts ...StartOp) {
	for _, fn := range opts {
		fn(s)
	}
}

// Support functions

func RequireEtcdCount(t *testing.T, setup *ProcessSetup, min int) {
	count := len(setup.Etcds)
	require.True(t, count >= min, "check minimum number of Etcds")
}

func RequireControllerCount(t *testing.T, setup *ProcessSetup, min int) {
	count := len(setup.Controllers)
	require.True(t, count >= min, "check minimum number of Controllers")
}

func RequireDmeCount(t *testing.T, setup *ProcessSetup, min int) {
	count := len(setup.Dmes)
	require.True(t, count >= min, "check minimum number of Dmes")
}

func RequireCrmCount(t *testing.T, setup *ProcessSetup, min int) {
	count := len(setup.Crms)
	require.True(t, count >= min, "check minimum number of Crms")
}

func ResetEtcds(t *testing.T, setup *ProcessSetup, count int) {
	for ii := 0; ii < count; ii++ {
		err := setup.Etcds[ii].ResetData()
		assert.Nil(t, err, "reset etcd ", ii)
	}
}

func StartEtcds(t *testing.T, setup *ProcessSetup, count int) {
	for ii := 0; ii < count; ii++ {
		err := setup.Etcds[ii].Start()
		require.Nil(t, err, "start etcd ", ii)
	}
}

func StopEtcds(setup *ProcessSetup, count int) {
	for ii := 0; ii < count; ii++ {
		setup.Etcds[ii].Stop()
	}
}

func StartControllers(t *testing.T, setup *ProcessSetup, count int, opts ...StartOp) {
	for ii := 0; ii < count; ii++ {
		err := setup.Controllers[ii].Start(opts...)
		require.Nil(t, err, "start controller ", ii)
	}
}

func StopControllers(setup *ProcessSetup, count int) {
	for ii := 0; ii < count; ii++ {
		setup.Controllers[ii].Stop()
	}
}

func ConnectControllerAPIs(t *testing.T, setup *ProcessSetup, count int, timeout time.Duration) []*grpc.ClientConn {
	conns := make([]*grpc.ClientConn, 0)
	for ii := 0; ii < count; ii++ {
		conn, err := setup.Controllers[ii].ConnectAPI(timeout)
		require.Nil(t, err, "connect controller API ", ii)
		conns = append(conns, conn)
	}
	return conns
}
