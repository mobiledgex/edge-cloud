package tasks

import (
	"context"
	"sync"
	"testing"

	"github.com/mobiledgex/edge-cloud/log"
	"github.com/stretchr/testify/require"
)

func TestKeyWorkers(t *testing.T) {
	log.InitTracer(nil)
	defer log.FinishTracer()
	ctx := log.StartTestSpan(context.Background())

	kw := KeyWorkers{}

	workdata := testWorkData{}
	// we use this chan to block the go thread so we can have
	// a deterministic count.
	workdata.startOk = make(chan bool, 1)
	workdata.data = make(map[interface{}]int)

	kw.Init("test", workdata.run)

	rep := 1000
	for ii := 0; ii < rep; ii++ {
		kw.NeedsWork(ctx, "key1")
		kw.NeedsWork(ctx, "key2")
		kw.NeedsWork(ctx, "key3")
	}
	close(workdata.startOk)

	kw.WaitIdle()

	require.Equal(t, 0, kw.WorkerCount())
	// these should all be 2, because the first run
	// gets stuck on startOk, and all susbsequent updates
	// are compressed into a single run.
	require.Equal(t, 2, workdata.data["key1"])
	require.Equal(t, 2, workdata.data["key2"])
	require.Equal(t, 2, workdata.data["key3"])
}

type testWorkData struct {
	startOk chan bool
	data    map[interface{}]int
	mux     sync.Mutex
}

func (s *testWorkData) run(ctx context.Context, key interface{}) {
	<-s.startOk

	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[key]++
}
