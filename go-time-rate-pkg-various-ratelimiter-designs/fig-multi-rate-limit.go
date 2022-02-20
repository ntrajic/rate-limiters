package main

import (
	"context"
	"log"
	"os"
	"sort"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// simple aggregate rate limiter called multiLimiter
// ------------------------------------------------
func main() {
	defer log.Printf("Done.")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
			log.Printf("ReadFile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
}
func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration / time.Duration(eventCount))
}

// redefine our APIConnection to have limits both per second and per minute:
func Open() *APIConnection {
	secondLimit := rate.NewLimiter(Per(2, time.Second), 1)   // <1> limit per second with no burstiness.
	minuteLimit := rate.NewLimiter(Per(10, time.Minute), 10) // <2> limit per minute with a burstiness of 10
	return &APIConnection{
		rateLimiter: MultiLimiter(secondLimit, minuteLimit), // <3> combine the two limits and set this as the master rate limiter for our APIConnection
	}
}

type APIConnection struct {
	rateLimiter RateLimiter
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

type RateLimiter interface { // <1> a RateLimiter interface so that a MultiLimiter can recursively define other MultiLimiter instances.
	Wait(context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit) // <2> implement an optimization and sort by the Limit() of each RateLimiter
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _, l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit() // <3> sort the child RateLimiter instances when multiLimiter is instantiated, so we can simply return the most restrictive limit, which will be the first element in the slice.
}

// NOTE1: we make two requests per second up until request #11,
// at which point we begin making requests every six seconds.
// NOTE2:
// It might be slightly counterintuitive why request #11
// occurs after only two seconds rather than six like the rest of the requests.
// Solution:
// Although we limit our API requests to 10 a minute,
// that minute is **a sliding window** of time.
// By the time we reach the eleventh request,
// our per-minute rate limiter has accrued another token.
//
// NOTE3: Multi-dimensional rate-limiting:
// This technique also allows us to begin thinking across dimensions
// other than time.
// When you rate limit a system, you’re probably going to limit more than one thing.
// - limit on the number of API requests,
// - limit other resources like: disk access, network access, etc.
//
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting>
// $ go run fig-multi-rate-limit.go
// OUT:
//                              ------------------
//1 08:03:09 ReadFile
//2 08:03:10 ReadFile                   ^
//3 08:03:10 ResolveAddress             |
//4 08:03:11 ReadFile                  each 1 sec
//5 08:03:11 ReadFile
//6 08:03:12 ReadFile
//7 08:03:12 ResolveAddress
//8 08:03:13 ResolveAddress
//0 08:03:13 ResolveAddress             |
//10 08:03:14 ResolveAddress            v
//11 08:03:15 ResolveAddress    ------------------
// 08:03:21 ReadFile                    ^
// 08:03:27 ReadFile
// 08:03:33 ResolveAddress              |
// 08:03:39 ReadFile
// 08:03:45 ReadFile                   each 6 sec
// 08:03:51 ReadFile
// 08:03:57 ResolveAddress              |
// 08:04:03 ResolveAddress
// 08:04:09 ResolveAddress              v
// 08:04:09 Done.              ---------------------

// NOTE:
// The Wait method loops through all the child rate limiters and calls Wait on each of them.
// These calls may or may not block, but we need to notify each rate limiter of the request
// so we can decrement our token bucket.
// By waiting for each limiter, we are guaranteed to wait for exactly the time of the longest wait.
