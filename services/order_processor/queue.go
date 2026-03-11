package main

import "sync"

type UnboundedQueue[T any] struct {
	mu   sync.Mutex
	cond *sync.Cond
	q    []T
}

func NewUnboundedQueue[T any]() *UnboundedQueue[T] {
	u := &UnboundedQueue[T]{q: make([]T, 0, 1024)}
	u.cond = sync.NewCond(&u.mu)
	return u
}

func (u *UnboundedQueue[T]) Enqueue(v T) {
	u.mu.Lock()
	u.q = append(u.q, v)
	u.mu.Unlock()
	u.cond.Signal()
}

func (u *UnboundedQueue[T]) DequeueBlocking() T {
	u.mu.Lock()
	for len(u.q) == 0 {
		u.cond.Wait()
	}
	v := u.q[0]
	u.q = u.q[1:]
	u.mu.Unlock()
	return v
}

func (u *UnboundedQueue[T]) Len() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	return len(u.q)
}

func (u *UnboundedQueue[T]) Peek() (T, bool) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if len(u.q) == 0 {
		var zero T
		return zero, false
	}

	return u.q[0], true
}
