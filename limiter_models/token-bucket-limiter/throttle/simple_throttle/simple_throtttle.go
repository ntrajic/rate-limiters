package simple_throttle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Effector func(context.Context) (string, error)

// basic “token bucket” algorithm that uses the “error” strategy:
func Throttle(e Effector, max uint, refill uint, d time.Duration) Effector {
	var tokens = max
	var once sync.Once
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		once.Do(func() {									// <---- once exec, inside passed anonym func() is lock/unlock
			ticker := time.NewTicker(d)                     // <---- ticker timer

			go func() {
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return

					case <-ticker.C:
						m.Lock()						//<----- LOCK
						t := tokens + refill
						if t > max {
							t = max
						}
						tokens = t
						m.Unlock()                      //<---- UNLOCK
					}
				}
			}()
		})

		m.Lock()
		defer m.Unlock()

		if tokens <= 0 {
			return "", fmt.Errorf("too many calls")
		}

		tokens--

		return e(ctx)
	}
}
// ==========================================
// simple_throttle rate-limitter algorithm:
// ==========================================
// - Throttle function wraps the effector function e with a closure 
// that contains the rate-limiting logic. 
// - The bucket is initially allocated max tokens; 
//   each time the closure is triggered it checks whether it has any remaining tokens. 
// - If tokens are available, it decrements the token count by one and 
//   triggers the effector function. 
// - If not, an error is returned. 
// - Tokens are added at a rate of refill tokens every duration d.
