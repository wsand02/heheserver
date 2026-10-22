package thumbgen

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDoCoalescesSameKey(t *testing.T) {
	group := New(2)
	const callers = 8
	var generated atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	var wg sync.WaitGroup
	errors := make(chan error, callers)
	values := make(chan any, callers)

	for range callers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := group.Do("same", func() (any, error) {
				generated.Add(1)
				close(started)
				<-release
				return "thumbnail", nil
			})
			values <- value
			errors <- err
		}()
	}

	<-started
	time.Sleep(25 * time.Millisecond)
	if count := generated.Load(); count != 1 {
		t.Fatalf("generation count while request is in flight = %d, want 1", count)
	}
	close(release)
	wg.Wait()
	close(values)
	close(errors)
	for err := range errors {
		if err != nil {
			t.Errorf("Do() error = %v", err)
		}
	}
	for value := range values {
		if value != "thumbnail" {
			t.Errorf("Do() value = %v, want thumbnail", value)
		}
	}
	if count := generated.Load(); count != 1 {
		t.Errorf("generation count = %d, want 1", count)
	}
}

func TestDoBoundsDistinctKeys(t *testing.T) {
	group := New(2)
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	var wg sync.WaitGroup

	for _, key := range []string{"one", "two", "three"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, _ = group.Do(key, func() (any, error) {
				started <- struct{}{}
				<-release
				return key, nil
			})
		}(key)
	}

	<-started
	<-started
	select {
	case <-started:
		t.Fatal("more than two distinct keys started concurrently")
	case <-time.After(25 * time.Millisecond):
	}

	close(release)
	<-started
	wg.Wait()
}
