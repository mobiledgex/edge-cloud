// Empty wrapper for now. Later will enforce lock ordering

package util

import "sync"

type Mutex struct {
	mux sync.Mutex
}

func (m *Mutex) Lock() {
	m.mux.Lock()
}

func (m *Mutex) Unlock() {
	m.mux.Unlock()
}

func (m *Mutex) InitCond(cond *sync.Cond) {
	cond.L = &m.mux
}
