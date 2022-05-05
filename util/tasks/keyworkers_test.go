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

package tasks

import (
	"context"
	"sync"
	"testing"

	"github.com/edgexr/edge-cloud/log"
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

	workdata.spawned.Add(3)
	kw.NeedsWork(ctx, "key1")
	kw.NeedsWork(ctx, "key2")
	kw.NeedsWork(ctx, "key3")
	// wait until work threads have started and "needsWork" is cleared
	workdata.spawned.Wait()
	// these next 1000 should be compressed into one more iteration only.
	rep := 1000
	for ii := 0; ii < rep; ii++ {
		workdata.spawned.Add(3)
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
	spawned sync.WaitGroup
}

func (s *testWorkData) run(ctx context.Context, key interface{}) {
	s.spawned.Done()
	<-s.startOk

	s.mux.Lock()
	defer s.mux.Unlock()
	s.data[key]++
}
