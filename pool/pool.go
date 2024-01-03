package pool

import "sync"

func New[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{New: func() any { return fn() }}}
}

func (p *Pool[T]) Get() T  { return p.p.Get().(T) }
func (p *Pool[T]) Put(x T) { p.p.Put(x) }

type Pool[T any] struct{ p sync.Pool }
