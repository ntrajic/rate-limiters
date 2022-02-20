package main

import (
	"context"
	"log"
	"os"
	"sync"

	// an implementation of a token bucket rate limiter
	// from the golang.org/x/time/rate package.
	"golang.org/x/time/rate"
)

// MAINLINE: 10 async reads and resolves

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
func Open() *APIConnection {
	return &APIConnection{
		rateLimiter: rate.NewLimiter(rate.Limit(1), 1), // <1> set the rate limit for all API connections to one event per second.
	}
}

type APIConnection struct {
	rateLimiter *rate.Limiter
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil { // <2> wait on the rate limiter to have enough access tokens for us to complete our request.
		return err
	}
	// Pretend we do work here
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil { // <2>
		return err
	}
	// Pretend we do work here
	return nil
}

// PREREQUISITE:
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting>
// $ go mod init rate-limit.com/m/v1 <enter>
// OUT:
// go: creating new go.mod: module rate-limit.com/m/v1
// go: to add module requirements and sums:
// go mod tidy
//
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting$ dir
// fig-multi-rate-limit.go  fig-no-rate-limit.go  fig-simple-rate-limit.go  fig-tiered-rate-limit.go  go.mod
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting>
// $ go mod tidy <enter>
// OUT:
// go: finding module for package golang.org/x/time/rate
// go: found golang.org/x/time/rate in golang.org/x/time v0.0.0-20220210224613-90d013bbcef8
//
// $ go run fig-simple-rate-limit.go <enter>
//
// OUT:
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting>
// $ go run fig-simple-rate-limit.go
// 07:39:49 ResolveAddress
// 07:39:50 ReadFile
// 07:39:51 ReadFile
// 07:39:52 ReadFile
// 07:39:53 ResolveAddress
// 07:39:54 ResolveAddress
// 07:39:55 ResolveAddress
// 07:39:56 ResolveAddress
// 07:39:57 ResolveAddress
// 07:39:58 ResolveAddress
// 07:39:59 ResolveAddress
// 07:40:00 ReadFile
// 07:40:01 ReadFile
// 07:40:02 ReadFile
// 07:40:03 ResolveAddress
// 07:40:04 ResolveAddress
// 07:40:05 ReadFile
// 07:40:06 ReadFile
// 07:40:07 ReadFile
// 07:40:08 ReadFile
// 07:40:08 Done.

//--------------------------
// TOKEN-BUCKET ALGORITHM:
//--------------------------
// To utilize a resource, you have to have **an access token** for the resource.
// Without the token, your request is **denied**.
// Now imagine these tokens are stored in a bucket waiting to be retrieved for usage.
// The bucket has a depth of d =>  bucket can hold d access tokens at a time.
// Every time you need to access a resource, you reach into the bucket and remove a token.
//
// E.g. If your bucket contains five tokens => d = 5
// and you access the resource five times, you’d be able to do so;
// but on the sixth try, no access token would be available.
//
// => You either have to queue your request until a token becomes available, or deny the request.
//
//     stack or queue
//     -----------
//     | 1 2 3 5 5   -----> pop()
//     -----------   <----- push(token)
//
//    In the token bucket algorithm:
//    r = rate at which tokens are added back to the bucket.
//        It can be one a nanosecond, or one a minute.
//    This becomes what we commonly think of as the rate limit:
//    because we have to wait until new tokens become available,
//    we limit our operations to that refresh rate.
//
//    d = # of tokens available for immediate use
//    r = the rate at which tokens are replenished / consumed
//    control:
//    1 rate_limit
//    2 burstiness =  how many requests can be made when the bucket is full
//
//    Assume: have access to an API, and a Go client has been provided to utilize it
//    the API - API has two endpoints:
//    1 one for reading a file, and
//    2 one for resolving a domain name to an IP address.
//    Return values that would be needed to actually access a service:
//
//    CLIENT:
//==============================================================================
//    func Open() *APIConnection {
//     return &APIConnection{}
// }
//
// type APIConnection struct {}
//
// func (a *APIConnection) ReadFile(ctx context.Context) error {
//     // Pretend we do work here
//     return nil
// }

// func (a *APIConnection) ResolveAddress(ctx context.Context) error {
//     // Pretend we do work here
//     return nil
// }
//==============================================================================
//  MAINLINE:
//  simple driver to access this API
//  The driver needs to read 10 files and resolve 10 addresses,
//  but the files and addresses have no relation to each other and
//  so the driver can make these API calls concurrent to one another.
//  ================================================================
//
//
// golang.org/x/time/rate package:
//
//	Limit==0 => no events allowed
//  type Limit float64   // maximum frequency of some events; #events/sec; L
//  func NewLimiter(r Limit, b int) *Limiter // r - rate b - bucket depth
//
//  rate limits in terms of the number of operations per time measurement, not the interval between requests:
//  rate.Limit(events/timePeriod.Seconds())
//  rate.Inf  — an indication that there is no limit — if the interval provided is zero.
//
//  Every func:
//  func Per(eventCount int, duration time.Duration) rate.Limit {
//     return rate.Every(duration/time.Duration(eventCount))
// }
//
// rate.Limiter:
// we’ll want to use it to block our requests until we’re given an access token.
// block request with the Wait method, which simply calls WaitN(1)
//
// Wait is shorthand for WaitN(ctx, 1).
// func (lim *Limiter) Wait(ctx context.Context)
//
// WaitN blocks until lim permits n events to happen.
// It returns an error if n exceeds the Limiter's burst size, the Context is
// canceled, or the expected wait time exceeds the Context's Deadline.
// func (lim *Limiter) WaitN(ctx context.Context, n int) (err error)
