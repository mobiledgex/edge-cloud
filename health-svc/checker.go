package main

import (
	"time"
)

// Checker is a generic thread that spawns other
// threads to do health checks periodically.

type Checker struct {
	target        Target
	sig           chan bool
	interval      time.Duration
	maxConcurrent int
}

type Target interface {
	TagAll()
	GetTagged() map[Tagged]struct{}
	CheckOne(t Tagged)
}

type Tagged interface{}

func (s *Checker) start(target Target, interval time.Duration, maxConcurrent int) {
	s.target = target
	s.interval = interval
	s.maxConcurrent = maxConcurrent
	s.sig = make(chan bool, 1)
	s.target.TagAll()
	// this thread runs the checks when triggered
	go func() {
		for {
			s.checkTagged()
			select {
			case <-s.sig:
			}
		}
	}()
	// this threads tags all objects for checking every interval
	// and triggers the checker
	go func() {
		for {
			select {
			case <-time.After(s.interval):
				s.target.TagAll()
			}
			s.wakeup()
		}
	}()
}

func (s *Checker) checkTagged() {
	tagged := s.target.GetTagged()

	// iterate over all tagged and check each.
	// run up to max concurrent in parallel until done.
	sem := make(chan bool, s.maxConcurrent)
	for tag, _ := range tagged {
		sem <- true
		go func(t Tagged) {
			s.target.CheckOne(t)
			<-sem
		}(tag)
	}
	// once we can fill up sem to max concurrent,
	// then all go funcs must be done
	for ii := 0; ii < cap(sem); ii++ {
		sem <- true
	}
}

func (s *Checker) wakeup() {
	select {
	case s.sig <- true:
	default:
	}
}
