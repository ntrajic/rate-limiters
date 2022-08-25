package simple_throttle

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// callsCountFunction returns a function that increments the int that
// callCounter points to whenever it is run.
func callsCountFunction(callCounter *int) Effector {
	return func(ctx context.Context) (string, error) {
		*callCounter++
		return fmt.Sprintf("call %d", *callCounter), nil
	}
}

// TestThrottleMax1 tests whether a max of 1 call per duration is respected.
func TestThrottleMax1(t *testing.T) {
	const max uint = 1

	callsCounter := 0
	effector := callsCountFunction(&callsCounter)

	ctx := context.Background()
	throttle := Throttle(effector, max, max, time.Second)

	for i := 0; i < 100; i++ {
		throttle(ctx)
	}

	if callsCounter == 0 {
		t.Error("test is broken; got", callsCounter)
	}

	if callsCounter > int(max) {
		t.Error("max is broken; got", callsCounter)
	}
}

// TestThrottleMax10 tests whether a max of 10 calls per duration is respected.
func TestThrottleMax10(t *testing.T) {
	const max uint = 10

	callsCounter := 0
	effector := callsCountFunction(&callsCounter)

	ctx := context.Background()
	throttle := Throttle(effector, max, max, time.Second)

	for i := 0; i < 100; i++ {
		throttle(ctx)
	}

	if callsCounter == 0 {
		t.Error("test is broken; got", callsCounter)
	}

	if callsCounter > int(max) {
		t.Error("max is broken; got", callsCounter)
	}
}

// TestThrottleCallFrequency5Seconds tests whether a Throttle with a max of 1
// and a duration of 1 second called every 250ms for 5 seconds will be called
// exactly 5 times.
func TestThrottleCallFrequency5Seconds(t *testing.T) {
	callsCounter := 0
	effector := callsCountFunction(&callsCounter)

	ctx := context.Background()
	throttle := Throttle(effector, 1, 1, time.Second)

	// make a call every 1/4 second for 5 seconds.
	tickCounts := 0
	ticker := time.NewTicker(250 * time.Millisecond).C

	for range ticker {
		tickCounts++

		s, e := throttle(ctx)
		if e != nil {
			t.Log("Error:", e)
		} else {
			t.Log("output:", s)
		}

		if tickCounts >= 20 {
			break
		}
	}

	if callsCounter != 5 {
		t.Error("expected 5; got", callsCounter)
	}
}

// TestThrottleVariableRefill
func TestThrottleVariableRefill(t *testing.T) {
	callsCounter := 0
	effector := callsCountFunction(&callsCounter)

	ctx := context.Background()
	throttle := Throttle(effector, 4, 2, 500*time.Millisecond)

	tickCounts := 0
	ticker := time.NewTicker(250 * time.Millisecond)
	timer := time.NewTimer(2 * time.Second)

time:
	for {
		select {
		case <-ticker.C:
			tickCounts++

			s, e := throttle(ctx)
			if e != nil {
				t.Log("Error:", e)
			} else {
				t.Log("output:", s)
			}
		case <-timer.C:
			break time
		}
	}

	if callsCounter != 8 {
		t.Error("expected 8; got", callsCounter)
	}
}

// TestThrottleContextTimeout tests whether a Throttle will return an error
// when its context is canceled.
func TestThrottleContextTimeout(t *testing.T) {
	callsCounter := 0
	effector := callsCountFunction(&callsCounter)

	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	throttle := Throttle(effector, 1, 1, time.Second)

	s, e := throttle(ctx)
	if e != nil {
		t.Error("unexpected error:", e)
	} else {
		t.Log("output:", s)
	}

	// Wait for timeout
	time.Sleep(300 * time.Millisecond)

	_, e = throttle(ctx)
	if e != nil {
		t.Log("got expected error:", e)
	} else {
		t.Error("didn't get expected error")
	}
}

// 1. Windows platform.
// 2. Run each individual test from Visual Code ide:
// =================================================
// 2.1
// Running tool: C:\Program Files\Go\bin\go.exe test -timeout 30s -run ^TestThrottleMax1$ simple_throttle
//
// === RUN   TestThrottleMax1
// --- PASS: TestThrottleMax1 (0.00s)
// PASS
// ok      simple_throttle 0.058s
//
//
// > Test run finished at 8/25/2022, 12:06:14 PM <
//
// 2.2
// Running tool: C:\Program Files\Go\bin\go.exe test -timeout 30s -run ^TestThrottleMax10$ simple_throttle
//
// === RUN   TestThrottleMax10
// --- PASS: TestThrottleMax10 (0.00s)
// PASS
// ok      simple_throttle 0.062s
//
//
// > Test run finished at 8/25/2022, 12:09:32 PM <
//
//
// 2.3
// Running tool: C:\Program Files\Go\bin\go.exe test -timeout 30s -run ^TestThrottleCallFrequency5Seconds$ simple_throttle
//
// === RUN   TestThrottleCallFrequency5Seconds
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:86: output: call 1
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:86: output: call 2
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:86: output: call 3
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:86: output: call 4
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:86: output: call 5
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:84: Error: too many calls
// --- PASS: TestThrottleCallFrequency5Seconds (5.01s)
// PASS
// ok      simple_throttle 5.063s
//
//
// > Test run finished at 8/25/2022, 12:10:41 PM <
//
// 2.4
// Running tool: C:\Program Files\Go\bin\go.exe test -timeout 30s -run ^TestThrottleVariableRefill$ simple_throttle
//
// === RUN   TestThrottleVariableRefill
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 1
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 2
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 3
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 4
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 5
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 6
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 7
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:121: output: call 8
// --- PASS: TestThrottleVariableRefill (2.01s)
// PASS
// ok      simple_throttle 2.077s
//
//
// > Test run finished at 8/25/2022, 12:12:07 PM <
//
// 2.5
// Running tool: C:\Program Files\Go\bin\go.exe test -timeout 30s -run ^TestThrottleContextTimeout$ simple_throttle
//
// === RUN   TestThrottleContextTimeout
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:148: output: call 1
// c:\SRC\GoLang\rate-limiters\limiter_models\token-bucket-limiter\throttle\simple_throttle\simple_throttle_test.go:156: got expected error: context deadline exceeded
// --- PASS: TestThrottleContextTimeout (0.31s)
// PASS
// ok      simple_throttle 0.364s
//
//
// > Test run finished at 8/25/2022, 12:13:29 PM <
