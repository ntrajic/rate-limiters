package main

import (
	"context"
	"log"
	"os"
	"sync"
)

// Case: no rate-limit implemented to apiConnection
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
	return &APIConnection{}
}

type APIConnection struct{}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	// Pretend we do work here
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	// Pretend we do work here
	return nil
}

// NOTE:
// =============
// RATE LIMITING
// =============
// constrains the number of times some kind of resource is accessed to some finite number per unit of time.
// The resource can be anything: API connections, disk reads/writes, network packets, errors.
// By rate limiting a system, you secure your system:
// you prevent entire classes of attack vectors against your system.
// For example:
// - attackers could fill up your service’s disk either with log messages or valid requests.
// - attackers could manipulate w/ log rotation: any record of the activity they could rotate out of the log
//   and into /dev/null
// - attempt to brute-force access to a resource
// - perform a distributed denial of service attack.
//
// Rate limits allow you to reason about the performance and stability of your system
// by preventing it from falling outside the boundaries you’ve already investigated.
// If you need to expand those boundaries, you can do so in a controlled manner
//
//
// We can see that all API requests are fielded almost simultaneously.
// We have **no** rate limiting set up and
// so our clients are free to access the system as frequently as they like.
//
// OUT:
// ntrajic@DESKTOP-6PK7L32:/mnt/c/src/GoLang/ConcurrencyGo/concurrency-at-scale/rate-limiting$
// go run fig-no-rate-limit.go
//
// 06:47:38 ResolveAddress
// 06:47:38 ReadFile
// 06:47:38 ReadFile
// 06:47:38 ResolveAddress
// 06:47:38 ResolveAddress
// 06:47:38 ResolveAddress
// 06:47:38 ResolveAddress
// 06:47:38 ResolveAddress
// 06:47:38 ReadFile
// 06:47:38 ReadFile
// 06:47:38 ReadFile
// 06:47:38 ReadFile
// 06:47:38 ReadFile
// 06:47:38 ResolveAddress
// 06:47:38 ReadFile
// 06:47:38 ResolveAddress
// 06:47:38 ResolveAddress
// 06:47:38 ReadFile
// 06:47:38 ResolveAddress
// 06:47:38 ReadFile
// 06:47:38 Done.
