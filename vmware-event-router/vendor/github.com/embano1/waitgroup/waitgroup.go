package waitgroup

import (
	"errors"
	"sync"
	"time"
)

// WaitGroup wraps sync.WaitGroup and adds a method WaitTimeout to abort waiting
// for long-running, blocked or leaked goroutines blocking Wait from the
// underlying WaitGroup. A caller might use this functionality to terminate a
// program in a bounded time.
type WaitGroup struct {
	sync.WaitGroup
}

// ErrTimeout is returned when the timeout in WaitTimeout is exceeded
var ErrTimeout = errors.New("timed out")

// WaitTimeout blocks until the WaitGroup counter is zero or when timeout is
// exceeded. It spawns an internal goroutine. In case of timeout exceeded the
// error ErrTimeout is returned and the internally spawned goroutine might leak
// if Wait never returns from the underlying WaitGroup.
//
// It is safe to call WaitTimeout concurrently but doing so might leak
// additional goroutines as described above.
func (wg *WaitGroup) WaitTimeout(timeout time.Duration) error {
	doneCh := make(chan struct{})
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case <-timer.C:
		return ErrTimeout
	case <-doneCh:
		return nil
	}
}
