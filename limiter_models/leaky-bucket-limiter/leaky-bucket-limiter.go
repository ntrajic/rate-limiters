package limiter_models

import (
	"sync"
	"time"
)

// LeakyBucketLimiter
type LeakyBucketLimiter struct {
	peakLevel       int
	currentLevel    int
	currentVelocity int
	lastTime        time.Time
	mutex           sync.Mutex
}

func NewLeakyBucketLimiter(peakLevel, currentVelocity int) *LeakyBucketLimiter {
	return &LeakyBucketLimiter{
		peakLevel:       peakLevel,
		currentVelocity: currentVelocity,
		lastTime:        time.Now(),
	}
}

func (l *LeakyBucketLimiter) TryAcquire() bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	now := time.Now()

	interval := now.Sub(l.lastTime)
	if interval >= time.Second {

		l.currentLevel = maxInt(0, l.currentLevel-int(interval/time.Second)*l.currentVelocity)
		l.lastTime = now
	}

	if l.currentLevel >= l.peakLevel {
		return false
	}

	l.currentLevel++
	return true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
