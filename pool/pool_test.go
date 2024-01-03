package pool

import (
	"fmt"
	"sync"
	"testing"
)

func TestPool(t *testing.T) {
	type foo struct {
		mu  sync.RWMutex
		bar map[string]int
	}

	total := &foo{bar: make(map[string]int)}

	p := New(func() *foo { return &foo{} })

	wg := &sync.WaitGroup{}

	iter := 10000
	for i := 0; i < iter; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			x := p.Get()
			defer p.Put(x)

			total.mu.Lock()
			total.bar[fmt.Sprintf("%p", x)]++
			total.mu.Unlock()

		}()
	}
	wg.Wait()
	l := len(total.bar)
	if l <= 1 || l == iter {
		t.Errorf("аллоцированных переменных должно быть не меньше 2 и не больше %d, но получилось %d\n", iter, l)
	}
}
