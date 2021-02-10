package wait

import (
	"sync"
	"time"
)

type Wait struct {
	sync.WaitGroup
}

// WaitWithTimeout 带超时时间的关闭
// return false 自然关闭
// return true 超时关闭
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		w.Wait()
		c <- struct{}{}
	}()

	select {
	case <-c:
		return false
	case <-time.After(timeout):
		return true
	}
}
