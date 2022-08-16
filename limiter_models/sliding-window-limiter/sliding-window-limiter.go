package limiter_models

import (
	"errors"
	"sync"
	"time"
)

// SlidingWindowLimiter
type SlidingWindowLimiter struct {
	limit        int           
	window       int64         
	smallWindow  int64         
	smallWindows int64         
	counters     map[int64]int 
	mutex        sync.Mutex    
}

func NewSlidingWindowLimiter(limit int, window, smallWindow time.Duration) (*SlidingWindowLimiter, error) {

	if window%smallWindow != 0 {
		return nil, errors.New("window cannot be split by integers")
	}

	return &SlidingWindowLimiter{
		limit:        limit,
		window:       int64(window),
		smallWindow:  int64(smallWindow),
		smallWindows: int64(window / smallWindow),
		counters:     make(map[int64]int),
	}, nil
}

func (l *SlidingWindowLimiter) TryAcquire() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	currentSmallWindow := time.Now().UnixNano() / l.smallWindow * l.smallWindow

	startSmallWindow := currentSmallWindow - l.smallWindow*(l.smallWindows-1)

	var count int
	for smallWindow, counter := range l.counters {
		if smallWindow < startSmallWindow {
			delete(l.counters, smallWindow)
		} else {
			count += counter
		}
	}

	if count >= l.limit {
		return false
	}

	l.counters[currentSmallWindow]++
	return true
}