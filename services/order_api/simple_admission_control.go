package main

import "sync/atomic"

// simpleAdmissionControl is a fixed-capacity, non-blocking admission gate.
type simpleAdmissionControl struct {
	slots chan struct{}
	inUse atomic.Int64
	limit int64
}

func newSimpleAdmissionControl(limit int) *simpleAdmissionControl {
	s := &simpleAdmissionControl{
		slots: make(chan struct{}, limit),
		limit: int64(limit),
	}
	for range limit {
		s.slots <- struct{}{}
	}
	return s
}

// tryAcquireSlot attempts to reserve one admission slot without blocking.
func (s *simpleAdmissionControl) tryAcquireSlot() bool {
	select {
	case <-s.slots:
		s.inUse.Add(1)
		return true
	default:
		return false
	}
}

// releaseSlot returns one slot to the pool.
func (s *simpleAdmissionControl) releaseSlot() {
	s.inUse.Add(-1)
	s.slots <- struct{}{}
}

func (s *simpleAdmissionControl) slotsInUse() int64 { return s.inUse.Load() }
func (s *simpleAdmissionControl) slotLimit() int64  { return s.limit }
