package thumbgen

import (
	"fmt"
	"sync"
	"sync/atomic"
)

const DefaultMaxConcurrent = 4

type call struct {
	done  chan struct{}
	value any
	err   error
}

type Group struct {
	mu    sync.Mutex
	calls map[string]*call
	sem   chan struct{}
}

func New(maxConcurrent int) *Group {
	if maxConcurrent < 1 {
		panic("maxConcurrent must be positive")
	}
	return &Group{
		calls: make(map[string]*call),
		sem:   make(chan struct{}, maxConcurrent),
	}
}

var shared atomic.Pointer[Group]

func init() {
	shared.Store(New(DefaultMaxConcurrent))
}

// Configure sets the process-wide generation limit at server startup.
// Existing calls retain their original group; do not use this for live resizing.
func Configure(maxConcurrent int) {
	shared.Store(New(maxConcurrent))
}

func Do(key string, fn func() (any, error)) (any, error) {
	return shared.Load().Do(key, fn)
}

func (g *Group) Do(key string, fn func() (any, error)) (value any, err error) {
	g.mu.Lock()
	if existing, ok := g.calls[key]; ok {
		g.mu.Unlock()
		<-existing.done
		return existing.value, existing.err
	}

	current := &call{done: make(chan struct{})}
	g.calls[key] = current
	g.mu.Unlock()

	acquired := false
	func() {
		defer func() {
			if acquired {
				<-g.sem
			}
			if recovered := recover(); recovered != nil {
				err = fmt.Errorf("thumbnail generation panic: %v", recovered)
			}
		}()

		g.sem <- struct{}{}
		acquired = true
		value, err = fn()
	}()

	g.mu.Lock()
	current.value = value
	current.err = err
	delete(g.calls, key)
	close(current.done)
	g.mu.Unlock()
	return value, err
}
