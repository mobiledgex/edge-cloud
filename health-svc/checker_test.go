package main

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestChecker(t *testing.T) {
	maxConcurrent := 20
	genCount := 1000

	tt := &TestTarget{}
	tt.init(genCount)
	tt.pause = true

	c := &Checker{}
	c.start(tt, 20*time.Second, maxConcurrent)
	// since we're paused, there should be 20 threads queued
	waitCount(t, maxConcurrent, &tt.runCount)
	require.Equal(t, 0, tt.finishedCount)
	// unpause and all of then should finish
	tt.pause = false
	tt.cond.Broadcast()
	waitCount(t, genCount, &tt.finishedCount)

	// run another cycle of all of them
	tt.TagAll()
	c.wakeup()
	waitCount(t, 2*genCount, &tt.finishedCount)
	require.Equal(t, maxConcurrent, tt.maxRunCount)
}

type TestTarget struct {
	tagged        map[Tagged]struct{}
	checked       map[TestKey]struct{}
	genCount      int
	runCount      int
	maxRunCount   int
	finishedCount int
	pause         bool
	mux           sync.Mutex
	cond          *sync.Cond
}

type TestKey struct {
	id int
}

func (s *TestTarget) init(genCount int) {
	s.tagged = make(map[Tagged]struct{})
	s.checked = make(map[TestKey]struct{})
	s.genCount = genCount
	s.cond = sync.NewCond(&s.mux)
}

func (s *TestTarget) TagAll() {
	s.mux.Lock()
	defer s.mux.Unlock()
	for ii := 0; ii < s.genCount; ii++ {
		key := TestKey{id: ii}
		s.tagged[key] = struct{}{}
	}
}

func (s *TestTarget) GetTagged() map[Tagged]struct{} {
	s.mux.Lock()
	defer s.mux.Unlock()
	tagged := s.tagged
	s.tagged = make(map[Tagged]struct{})
	return tagged
}

func (s *TestTarget) CheckOne(t Tagged) {
	key, ok := t.(TestKey)
	if !ok {
		panic("invalid type")
	}
	s.mux.Lock()
	s.checked[key] = struct{}{}
	s.runCount++
	if s.runCount > s.maxRunCount {
		s.maxRunCount = s.runCount
	}
	for s.pause {
		s.cond.Wait()
	}
	s.runCount--
	s.finishedCount++
	s.mux.Unlock()
}

func waitCount(t *testing.T, expected int, counter *int) {
	for i := 0; i < 10; i++ {
		if expected == *counter {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.Equal(t, expected, *counter)
}
