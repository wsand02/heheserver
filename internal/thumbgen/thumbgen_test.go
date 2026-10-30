package thumbgen

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wsand02/heheserver/internal/resize"
)

func benchmarkResizeSource(b *testing.B) []byte {
	b.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 600, 400))
	for y := 0; y < 400; y++ {
		for x := 0; x < 600; x++ {
			img.Set(x, y, color.RGBA{uint8(x % 256), uint8(y % 256), 64, 255})
		}
	}
	var source bytes.Buffer
	if err := png.Encode(&source, img); err != nil {
		b.Fatal(err)
	}
	return source.Bytes()
}

func BenchmarkResizeWorkerGroup(b *testing.B) {
	source := benchmarkResizeSource(b)
	keys := make([]string, 1<<16)
	for i := range keys {
		keys[i] = strconv.Itoa(i)
	}

	for _, maxConcurrent := range []int{0, 1, 4, 16} {
		name := "NoGroup"
		var group *Group
		if maxConcurrent > 0 {
			name = "Workers" + strconv.Itoa(maxConcurrent)
			group = New(maxConcurrent)
		}
		b.Run(name, func(b *testing.B) {
			var nextKey atomic.Uint64
			b.ReportAllocs()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					if group == nil {
						if _, err := resize.ResizeImageFallback(bytes.NewReader(source)); err != nil {
							b.Error(err)
						}
						continue
					}
					key := keys[nextKey.Add(1)&uint64(len(keys)-1)]
					_, err := group.Do(key, func() (any, error) {
						return resize.ResizeImageFallback(bytes.NewReader(source))
					})
					if err != nil {
						b.Error(err)
					}
				}
			})
		})
	}
}

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

func TestConfigureSharedLimit(t *testing.T) {
	Configure(1)
	t.Cleanup(func() { Configure(DefaultMaxConcurrent) })
	started := make(chan string, 2)
	release := make(chan struct{})
	var wg sync.WaitGroup
	for _, key := range []string{"image:test", "video:test"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			_, _ = Do(key, func() (any, error) {
				started <- key
				<-release
				return nil, nil
			})
		}(key)
	}
	<-started
	select {
	case <-started:
		t.Error("configured limit exceeded across image and video jobs")
	case <-time.After(25 * time.Millisecond):
	}
	close(release)
	wg.Wait()
}

func TestDoReleasesSlotAndKeyAfterFailure(t *testing.T) {
	group := New(1)
	for _, panicJob := range []bool{false, true} {
		_, err := group.Do("same", func() (any, error) {
			if panicJob {
				panic("test panic")
			}
			return nil, errors.New("test failure")
		})
		if err == nil {
			t.Fatal("expected failure")
		}
		done := make(chan struct{})
		go func() {
			defer close(done)
			value, err := group.Do("same", func() (any, error) { return "recovered", nil })
			if err != nil || value != "recovered" {
				t.Errorf("retry: value=%v err=%v", value, err)
			}
		}()
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("failure leaked slot or in-flight key")
		}
	}
}
