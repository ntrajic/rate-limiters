package limiter_models

import (
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ViolationStrategyError
type ViolationStrategyError struct {
	Limit  int
	Window time.Duration
}

func (e *ViolationStrategyError) Error() string {
	return fmt.Sprintf("violation strategy that limit = %d and window = %d", e.Limit, e.Window)
}

// SlidingLogLimiterStrategy
type SlidingLogLimiterStrategy struct {
	limit        int
	window       int64
	smallWindows int64
}

func NewSlidingLogLimiterStrategy(limit int, window time.Duration) *SlidingLogLimiterStrategy {
	return &SlidingLogLimiterStrategy{
		limit:  limit,
		window: int64(window),
	}
}

// SlidingLogLimiter
type SlidingLogLimiter struct {
	strategies  []*SlidingLogLimiterStrategy
	smallWindow int64
	counters    map[int64]int
	mutex       sync.Mutex
}

func NewSlidingLogLimiter(smallWindow time.Duration, strategies ...*SlidingLogLimiterStrategy) (*SlidingLogLimiter, error) {

	strategies = append(make([]*SlidingLogLimiterStrategy, 0, len(strategies)), strategies...)

	if len(strategies) == 0 {
		return nil, errors.New("must be set strategies")
	}

	sort.Slice(strategies, func(i, j int) bool {
		a, b := strategies[i], strategies[j]
		if a.window == b.window {
			return a.limit > b.limit
		}
		return a.window > b.window
	})

	for i, strategy := range strategies {

		if i > 0 {
			if strategy.limit >= strategies[i-1].limit {
				return nil, errors.New("the smaller window should be the smaller limit")
			}
		}

		if strategy.window%int64(smallWindow) != 0 {
			return nil, errors.New("window cannot be split by integers")
		}
		strategy.smallWindows = strategy.window / int64(smallWindow)
	}

	return &SlidingLogLimiter{
		strategies:  strategies,
		smallWindow: int64(smallWindow),
		counters:    make(map[int64]int),
	}, nil
}

func (l *SlidingLogLimiter) TryAcquire() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	currentSmallWindow := time.Now().UnixNano() / l.smallWindow * l.smallWindow

	startSmallWindows := make([]int64, len(l.strategies))
	for i, strategy := range l.strategies {
		startSmallWindows[i] = currentSmallWindow - l.smallWindow*(strategy.smallWindows-1)
	}

	counts := make([]int, len(l.strategies))
	for smallWindow, counter := range l.counters {
		if smallWindow < startSmallWindows[0] {
			delete(l.counters, smallWindow)
			continue
		}
		for i := range l.strategies {
			if smallWindow >= startSmallWindows[i] {
				counts[i] += counter
			}
		}
	}

	for i, strategy := range l.strategies {
		if counts[i] >= strategy.limit {
			return &ViolationStrategyError{
				Limit:  strategy.limit,
				Window: time.Duration(strategy.window),
			}
		}
	}

	l.counters[currentSmallWindow]++
	return nil
}
