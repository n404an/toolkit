package gsm // Package gsm generic sync map

import (
	"github.com/n404an/toolkit/pool"
	"sync"
	"sync/atomic"
	"unsafe"
)

func NewMap[K comparable, V any]() *Map[K, V] { return &Map[K, V]{} }

func NewMapOnPool[K comparable, V any]() *pool.Pool[*Map[K, V]] {
	return pool.New(func() *Map[K, V] { return &Map[K, V]{} })
}

func (m *Map[K, V]) Store(key K, value V)              { m.swap(key, value) }
func (m *Map[K, V]) Load(key K) (value V, ok bool)     { return m.loadTryRead(key) }
func (m *Map[K, V]) Delete(key K)                      { m.loadAndDelete(key) }
func (m *Map[K, V]) Range(f func(key K, value V) bool) { m.rng(f) }
func (m *Map[K, V]) Clear()                            { m.clear() }
func (m *Map[K, V]) Len() (count int)                  { return m.len() }

func (m *Map[K, V]) Swap(key K, value V) (previous V, loaded bool) {
	return m.swap(key, value)
}

func (m *Map[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.compareAndDelete(key, old)
}

func (m *Map[K, V]) CompareAndSwap(key K, old, new V) bool {
	return m.compareAndSwap(key, old, new)
}

func (m *Map[K, V]) LoadAndDelete(key K) (value V, loaded bool) {
	return m.loadAndDelete(key)
}

func (m *Map[K, V]) LoadOrStore(key K, value V) (actual any, loaded bool) {
	return m.loadOrStore(key, value)
}

type Map[K comparable, V any] struct {
	mu     sync.Mutex
	read   atomic.Pointer[readOnly[K, V]]
	dirty  map[K]*entry[V]
	misses int
}

type readOnly[K comparable, V any] struct {
	m       map[K]*entry[V]
	amended bool
}

var expunged = unsafe.Pointer(new(any))

type entry[V any] struct{ p atomic.Pointer[V] }

func newEntry[V any](i V) *entry[V] {
	e := &entry[V]{}
	e.p.Store(&i)
	return e
}

func (m *Map[K, V]) loadReadOnly() readOnly[K, V] {
	if p := m.read.Load(); p != nil {
		return *p
	}
	return readOnly[K, V]{}
}

func (m *Map[K, V]) loadTryRead(key K) (value V, ok bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			m.missLocked()
		}
		m.mu.Unlock()
	}
	if !ok {
		var zero V
		return zero, false
	}
	return e.load()
}

func (e *entry[V]) load() (value V, ok bool) {
	p := e.p.Load()
	if p == nil || unsafe.Pointer(p) == expunged {
		var zero V
		return zero, false
	}
	return *p, true
}

func (e *entry[V]) tryCompareAndSwap(old, new V) bool {
	p := e.p.Load()
	if p == nil || unsafe.Pointer(p) == expunged {
		return false
	}

	nc := new
	for {
		if e.p.CompareAndSwap(p, &nc) {
			return true
		}
		p = e.p.Load()
		if p == nil || unsafe.Pointer(p) == expunged {
			return false
		}
	}
}

func (e *entry[V]) unexpungeLocked() (wasExpunged bool) {
	return e.p.CompareAndSwap((*V)(expunged), nil)
}

func (e *entry[V]) swapLocked(i *V) *V { return e.p.Swap(i) }

func (m *Map[K, V]) loadOrStore(key K, value V) (actual any, loaded bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		actual, loaded, ok := e.tryLoadOrStore(value)
		if ok {
			return actual, loaded
		}
	}

	m.mu.Lock()
	read = m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {
			m.dirty[key] = e
		}
		actual, loaded, _ = e.tryLoadOrStore(value)
	} else if e, ok := m.dirty[key]; ok {
		actual, loaded, _ = e.tryLoadOrStore(value)
		m.missLocked()
	} else {
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(&readOnly[K, V]{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
		actual, loaded = value, false
	}
	m.mu.Unlock()

	return actual, loaded
}

func (e *entry[V]) tryLoadOrStore(i V) (actual V, loaded, ok bool) {
	p := e.p.Load()
	if unsafe.Pointer(p) == expunged {
		var zero V
		return zero, false, false
	}
	if p != nil {
		return *p, true, true
	}

	ic := i
	for {
		if e.p.CompareAndSwap(nil, &ic) {
			return i, false, true
		}
		p = e.p.Load()
		if unsafe.Pointer(p) == expunged {
			var zero V
			return zero, false, false
		}
		if p != nil {
			return *p, true, true
		}
	}
}
func (m *Map[K, V]) loadAndDelete(key K) (value V, loaded bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]
			delete(m.dirty, key)

			m.missLocked()
		}
		m.mu.Unlock()
	}
	if ok {
		return e.delete()
	}

	var zero V
	return zero, false
}

func (e *entry[V]) delete() (value V, ok bool) {
	for {
		p := e.p.Load()
		if p == nil || unsafe.Pointer(p) == expunged {
			var zero V
			return zero, false
		}
		if e.p.CompareAndSwap(p, nil) {
			return *p, true
		}
	}
}

func (e *entry[V]) trySwap(i *V) (*V, bool) {
	for {
		p := e.p.Load()
		if unsafe.Pointer(p) == expunged {
			return nil, false
		}
		if e.p.CompareAndSwap(p, i) {
			return p, true
		}
	}
}

func (m *Map[K, V]) swap(key K, value V) (previous V, loaded bool) {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if v, ok := e.trySwap(&value); ok {
			if v == nil {
				var zero V
				return zero, false
			}
			return *v, true
		}
	}

	m.mu.Lock()
	read = m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		if e.unexpungeLocked() {

			m.dirty[key] = e
		}
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else if e, ok := m.dirty[key]; ok {
		if v := e.swapLocked(&value); v != nil {
			loaded = true
			previous = *v
		}
	} else {
		if !read.amended {
			m.dirtyLocked()
			m.read.Store(&readOnly[K, V]{m: read.m, amended: true})
		}
		m.dirty[key] = newEntry(value)
	}
	m.mu.Unlock()
	return previous, loaded
}

func (m *Map[K, V]) compareAndSwap(key K, old, new V) bool {
	read := m.loadReadOnly()
	if e, ok := read.m[key]; ok {
		return e.tryCompareAndSwap(old, new)
	} else if !read.amended {
		return false
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	read = m.loadReadOnly()
	swapped := false
	if e, ok := read.m[key]; ok {
		swapped = e.tryCompareAndSwap(old, new)
	} else if e, ok := m.dirty[key]; ok {
		swapped = e.tryCompareAndSwap(old, new)

		m.missLocked()
	}
	return swapped
}

func (m *Map[K, V]) compareAndDelete(key K, old V) (deleted bool) {
	read := m.loadReadOnly()
	e, ok := read.m[key]
	if !ok && read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		e, ok = read.m[key]
		if !ok && read.amended {
			e, ok = m.dirty[key]

			m.missLocked()
		}
		m.mu.Unlock()
	}
	//	sync.Map{}
	for ok {
		p := e.p.Load()
		if p == nil || unsafe.Pointer(p) == expunged {
			return false
		}
		if e.p.CompareAndSwap(p, nil) {
			return true
		}
	}
	return false
}

func (m *Map[K, V]) rng(f func(key K, value V) bool) {
	read := m.loadReadOnly()
	if read.amended {
		m.mu.Lock()
		read = m.loadReadOnly()
		if read.amended {
			read = readOnly[K, V]{m: m.dirty}
			m.read.Store(&read)
			m.dirty = nil
			m.misses = 0
		}
		m.mu.Unlock()
	}

	for k, e := range read.m {
		v, ok := e.load()
		if !ok {
			continue
		}
		if !f(k, v) {
			break
		}
	}
}

func (m *Map[K, V]) missLocked() {
	m.misses++
	if m.misses < len(m.dirty) {
		return
	}
	m.read.Store(&readOnly[K, V]{m: m.dirty})
	m.dirty = nil
	m.misses = 0
}

func (m *Map[K, V]) dirtyLocked() {
	if m.dirty != nil {
		return
	}

	read := m.loadReadOnly()
	m.dirty = make(map[K]*entry[V], len(read.m))
	for k, e := range read.m {
		if !e.tryExpungeLocked() {
			m.dirty[k] = e
		}
	}
}

func (m *Map[K, V]) clear() {
	m.rng(func(key K, _ V) bool {
		m.Delete(key)
		return true
	})
}

func (m *Map[K, V]) len() (count int) {
	m.rng(func(_ K, _ V) bool {
		count++
		return true
	})
	return
}

func (e *entry[V]) tryExpungeLocked() (isExpunged bool) {
	p := e.p.Load()
	for p == nil {
		if e.p.CompareAndSwap(nil, (*V)(expunged)) {
			return true
		}
		p = e.p.Load()
	}
	return unsafe.Pointer(p) == expunged
}
