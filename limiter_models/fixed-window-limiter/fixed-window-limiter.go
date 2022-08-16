package limiter_models

import (
	"sync"
	"time"
)

// FixedWindowLimiter
type FixedWindowLimiter struct {
	limit    int            
	window   time.Duration  
	counter  int            
	lastTime time.Time      
	mutex    sync.Mutex     
}

func NewFixedWindowLimiter(limit int, window time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		limit:    limit,
		window:   window,
		lastTime: time.Now(),
	}
}

func (l *FixedWindowLimiter) TryAcquire() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	now := time.Now()
	if now.Sub(l.lastTime) > l.window {
		l.counter = 0
		l.lastTime = now
	}
	if l.counter >= l.limit {
		return false
	}
	l.counter++
	return true
}