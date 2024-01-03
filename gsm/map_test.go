package gsm

import (
	"strconv"
	"sync"
	"testing"
	"unsafe"
)

/*
cpu: AMD Ryzen 9 5950X 16-Core Processor
set_base_map_mu-32                     1214   1052427 ns/op      562 B/op       0 allocs/op
set_base_map_mu_batch-32               7080    168270 ns/op       95 B/op       0 allocs/op
set_sync_map-32                       18033     62538 ns/op    78106 B/op    9745 allocs/op
set_gen_sync_map-32                  117394      9604 ns/op       16 B/op       0 allocs/op
set_gen_sync_map_pool-32             113395      9640 ns/op       15 B/op       0 allocs/op
range_get_base_map_mu_batch-32       242991      4654 ns/op        0 B/op       0 allocs/op
range_get_sync_map-32                164001      7414 ns/op       16 B/op       1 allocs/op
range_get_gen_sync_map-32            201528      5731 ns/op       16 B/op       1 allocs/op
get_base_map_mu-32                 27519247     42.92 ns/op        0 B/op       0 allocs/op
get_sync_map-32                   397260206     3.728 ns/op        0 B/op       0 allocs/op
get_gen_sync_map-32               459431486     2.556 ns/op        0 B/op       0 allocs/op
*/

func TestGenericSyncMap(t *testing.T) {
	type tStruct struct {
		i int
	}

	m := NewMap[string, *tStruct]()
	cnt := 10000
	for i := 0; i < cnt; i++ {
		m.Store(strconv.Itoa(i), &tStruct{i})
	}

	key := cnt / 2
	prev, ok := m.Swap(strconv.Itoa(key), &tStruct{key * 2})
	if !ok {
		t.Error("отсутстует ключ", key)
	}
	if prev.i != key {
		t.Error("неправильное значение ", prev)
	}
	val, ok := m.Load(strconv.Itoa(key))
	if !ok {
		t.Error("отсутстует ключ", key)
	}
	if val.i != key*2 {
		t.Error("неправильное значение ", val)
	}
}

func TestGenericSyncMapCheck(t *testing.T) {
	type tStruct struct {
		i int
	}
	bMap := make(map[int]*tStruct)
	m := NewMap[int, *tStruct]()
	cnt := 10000

	for i := 0; i < cnt; i++ {
		bMap[i] = &tStruct{i}
	}

	for k, v := range bMap {
		m.Store(k, v)
	}

	if len(bMap) != m.Len() {
		t.Errorf("разные длины %d %d\n", len(bMap), m.Len())
	}

	for k, v := range bMap {
		val, ok := m.loadAndDelete(k)
		if !ok {
			t.Error("отсутстует ключ", k)
		}
		if unsafe.Pointer(v) != unsafe.Pointer(val) {
			t.Errorf("отличаются указатели %p %p %#v %#v\n", v, val, v, val)
		}
	}
	if m.Len() != 0 {
		t.Errorf("длина должна == 0, но на деле %d\n", m.Len())
	}
}

func BenchmarkMap(b *testing.B) {
	type tStruct struct {
		i int
	}
	type mStruct struct {
		mu sync.RWMutex
		m  map[int]*tStruct
	}
	totalKeys := 1000
	key := totalKeys / 2
	val := &tStruct{}
	b.Run("set base map mu", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		m := &mStruct{m: make(map[int]*tStruct)}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for i := 0; i < totalKeys; i++ {
					m.mu.Lock()
					m.m[i] = val
					m.mu.Unlock()
				}
			}
		})
	})
	b.Run("set base map mu batch", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		m := &mStruct{m: make(map[int]*tStruct)}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.mu.Lock()
				for i := 0; i < totalKeys; i++ {
					m.m[i] = val
				}
				m.mu.Unlock()
			}
		})
	})

	b.Run("set sync map", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		m := &sync.Map{}
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for i := 0; i < totalKeys; i++ {
					m.Store(i, val)
				}
			}
		})
	})

	b.Run("set gen sync map", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		m := NewMap[int, *tStruct]()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for i := 0; i < totalKeys; i++ {
					m.Store(i, val)
				}
			}
		})
	})
	b.Run("set gen sync map pool", func(b *testing.B) {
		p := NewMapOnPool[int, *tStruct]()
		b.ReportAllocs()
		b.ResetTimer()

		m := p.Get()
		m.Clear()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				for i := 0; i < totalKeys; i++ {
					m.Store(i, val)
				}
			}
		})
	})

	b.Run("range get base map mu batch", func(b *testing.B) {
		m := &mStruct{m: make(map[int]*tStruct)}
		for i := 0; i < totalKeys; i++ {
			m.m[i] = &tStruct{i}
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.mu.RLock()
				for k, v := range m.m {
					_ = k + v.i
				}
				m.mu.RUnlock()
			}
		})
	})
	b.Run("range get sync map", func(b *testing.B) {
		m := &sync.Map{}
		for i := 0; i < totalKeys; i++ {
			m.Store(i, &tStruct{i})
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Range(func(k any, v any) bool {
					_ = k.(int) + v.(*tStruct).i
					return true
				})
			}
		})
	})

	b.Run("range get gen sync map", func(b *testing.B) {
		m := NewMap[int, *tStruct]()
		for i := 0; i < totalKeys; i++ {
			m.Store(i, &tStruct{i})
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Range(func(k int, v *tStruct) bool {
					_ = k + v.i
					return true
				})
			}
		})

	})

	b.Run("get base map mu", func(b *testing.B) {
		m := &mStruct{m: make(map[int]*tStruct)}
		for i := 0; i < totalKeys; i++ {
			m.m[i] = &tStruct{i}
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.mu.RLock()
				_ = m.m[key]
				m.mu.RUnlock()
			}
		})
	})
	b.Run("get sync map", func(b *testing.B) {
		m := &sync.Map{}
		for i := 0; i < totalKeys; i++ {
			m.Store(i, &tStruct{i})
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Load(key)
			}
		})
	})

	b.Run("get gen sync map", func(b *testing.B) {
		m := NewMap[int, *tStruct]()
		for i := 0; i < totalKeys; i++ {
			m.Store(i, &tStruct{i})
		}
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				m.Load(key)
			}
		})
	})
}
