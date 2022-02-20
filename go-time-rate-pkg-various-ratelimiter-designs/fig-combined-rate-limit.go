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

func Open() *APIConnection {
	return &APIConnection{
		apiLimit: MultiLimiter( // 1 set up a rate limiter for API calls. There are limits for both requests per second and requests per minute.
			rate.NewLimiter(Per(2, time.Second), 2),
			rate.NewLimiter(Per(10, time.Minute), 10),
		),
		diskLimit: MultiLimiter( // 2 set up a rate limiter for disk reads. We’ll only limit this to one read per second.
			rate.NewLimiter(rate.Limit(1), 1),
		),
		networkLimit: MultiLimiter( // 3 set up a network limit of three requests per second.
			rate.NewLimiter(Per(3, time.Second), 3),
		),
	}
}

type APIConnection struct {
	networkLimit,
	diskLimit,
	apiLimit RateLimiter
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	err := MultiLimiter(a.apiLimit, a.diskLimit).Wait(ctx) //4 When we go to read a file, we’ll combine the limits from the API limiter and the disk limiter.
	if err != nil {
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	err := MultiLimiter(a.apiLimit, a.networkLimit).Wait(ctx) //5 When we require network access, we’ll combine the limits from the API limiter and the network limiter.
	if err != nil {
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

// OUT:
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting>
// $ go run fig-combined-rate-limit.go 08:38:07 ResolveAddress
// 08:38:07 ReadFile
// 08:38:08 ResolveAddress
// 08:38:08 ReadFile
// 08:38:09 ResolveAddress
// 08:38:09 ResolveAddress
// 08:38:10 ResolveAddress
// 08:38:10 ReadFile
// 08:38:11 ReadFile
// 08:38:12 ReadFile
// 08:38:13 ResolveAddress
// 08:38:19 ResolveAddress
// 08:38:25 ResolveAddress
// 08:38:31 ResolveAddress
// 08:38:37 ResolveAddress
// 08:38:43 ReadFile
// 08:38:49 ReadFile
// 08:38:55 ReadFile
// 08:39:01 ReadFile
// 08:39:07 ReadFile
// 08:39:07 Done.
