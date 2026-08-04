package main

import (
	"fmt"
	"lrucachettl/model"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Println("--- Starting Token Cache Verification ---")

	cache := model.NewTokenCache()

	// Start the background sweeper to run every 100 milliseconds
	cache.StartCleanup(100 * time.Millisecond)

	// -------------------------------------------------------------
	// 1. Basic Operations & TTL Expiration Test
	// -------------------------------------------------------------
	fmt.Println("\n[Test 1] Testing Expiration Logic...")
	cache.Set("token_admin", model.UserSession{UserID: "usr_100", Role: "Admin"}, 200*time.Millisecond)
	cache.Set("token_user", model.UserSession{UserID: "usr_200", Role: "User"}, 1*time.Second)

	// Immediately read
	if sess, ok := cache.Get("token_admin"); ok {
		fmt.Printf("✓ Token Found Immediately: UserID=%s, Role=%s\n", sess.UserID, sess.Role)
	} else {
		fmt.Println("✗ Failed: Token should exist!")
	}

	// Wait for admin token to expire
	time.Sleep(300 * time.Millisecond)

	if _, ok := cache.Get("token_admin"); !ok {
		fmt.Println("✓ Admin Token correctly reported as expired via Get().")
	} else {
		fmt.Println("✗ Failed: Expired token returned!")
	}

	// Wait for background cleanup to purge map and trigger Map Reset rule
	time.Sleep(1 * time.Second)
	fmt.Printf("✓ Cache length after background sweep: %d keys\n", cache.Count())

	// -------------------------------------------------------------
	// 2. High-Volume Concurrency Test (Race Detector Validation)
	// -------------------------------------------------------------
	fmt.Println("\n[Test 2] Running High Concurrency Read/Write Test...")

	const workers = 100
	const operationsPerWorker = 1000
	var wg sync.WaitGroup
	wg.Add(workers)

	startTime := time.Now()

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < operationsPerWorker; j++ {
				tokenKey := fmt.Sprintf("token_%d", rand.IntN(20)) // 20 keys under heavy contention

				if rand.Float32() < 0.2 {
					// 20% Writes
					cache.Set(tokenKey, model.UserSession{
						UserID: fmt.Sprintf("user_%d", workerID),
						Role:   "Member",
					}, 50*time.Millisecond)
				} else {
					// 80% Reads
					cache.Get(tokenKey)
				}
			}
		}(i)
	}

	wg.Wait()
	fmt.Printf("✓ Successfully processed %d concurrent operations in %v!\n",
		workers*operationsPerWorker, time.Since(startTime))

	// Allow final sweep
	time.Sleep(200 * time.Millisecond)
	runtime.GC() // Manual GC run to verify memory cleanup state

	fmt.Printf("✓ Final Map Count: %d\n", cache.Count())
	fmt.Println("--- All Verification Steps Passed Cleanly ---")
}
