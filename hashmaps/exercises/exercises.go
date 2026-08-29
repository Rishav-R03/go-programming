package exercises

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"strings"
	"sync"
)

/*
1. The "Shrinking" Memory Leak
Problem: Write a program that creates a map, inserts 1,000,000 large strings (e.g., 1KB each), and then deletes all of them.
Measure the memory usage (using runtime.MemStats) before insertion, after insertion, and after deletion.
Challenge: You will notice memory does not drop significantly after deletion.
Modify the code to actually release the memory back to the OS without restarting the program. Key Concept: Maps grow but never shrink. Learn when to re-initialize (m = make(...)) to reclaim memory.
*/
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func generateRandomString(sizeInBytes int) string {
	var sb strings.Builder
	sb.Grow(sizeInBytes)
	for i := 0; i < sizeInBytes; i++ {
		sb.WriteByte(charset[rand.IntN(len(charset))])
	}
	return sb.String()
}

func printMemStats(label string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%-20s -> Alloc: %4d MB | Sys: %4d MB\n", label, m.Alloc/1024/1024, m.Sys/1024/1024)
}

func ExerciseOneFixed() {
	printMemStats("Before Allocation")

	heavyMap := make(map[int]string)
	for i := range 1000000 {
		heavyMap[i] = generateRandomString(1024)
	}
	printMemStats("After Allocation")

	// 1. Clear keys (Map bucket memory still stays allocated)
	for k := range heavyMap {
		delete(heavyMap, k)
	}
	printMemStats("After Delete Loop")

	// 2. KEY FIX: Dereference the map so garbage collector can reclaim bucket memory
	heavyMap = nil

	// 3. Force GC run (for demonstration purposes)
	runtime.GC()
	printMemStats("After Re-init + GC")
}

/*
2. The Concurrent Counter (Race Condition)
Problem: Create a map counts := make(map[string]int).
Launch 100 goroutines, each incrementing a specific key (e.g., "user_1") 1,000 times. Run the program with go run -race main.go. Challenge: The program will panic or report a data race. Fix this in two ways:

Using sync.RWMutex.
Using sync.Map (compare the performance and complexity).
Key Concept: Understanding that standard maps are not thread-safe and the trade-offs between mutexes and sync.Map.

*/

func ExerciseTwoConcurrencyMap() {
	map1 := make(map[string]int)
	map1["lottery"] = 100

	var wg sync.WaitGroup
	const workers = 100
	wg.Add(workers)

	for worker := 0; worker < workers; worker++ {
		go func() {
			defer wg.Done()

			// Unsafe map read/write across multiple goroutines
			val := map1["lottery"]
			val++
			map1["lottery"] = val
		}()
	}

	wg.Wait()
	fmt.Printf("Current value: %d\n", map1["lottery"])
}

// Solving race condition using sync.RWMutex

type SafeMap struct {
	mu sync.RWMutex
	m  map[string]int
}

func RWMutexSolution() {
	sm := SafeMap{m: make(map[string]int)}
	sm.m["lottery"] = 100

	const workers = 100

	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			sm.mu.Lock()
			val := sm.m["lottery"]
			val++
			sm.m["lottery"] = val
			sm.mu.Unlock()
		}()
	}
	wg.Wait()

	sm.mu.RLock()
	finalVal := sm.m["lottery"]
	sm.mu.RUnlock()

	fmt.Println("Final value ", finalVal)
}

func SyncMapSolution() {
	var smap sync.Map
	smap.Store("lottery", 100)

	const workers = 100
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				oldVal, loaded := smap.Load("lottery")
				if !loaded {
					return
				}
				cur := oldVal.(int)
				newVal := cur + 1

				if smap.CompareAndSwap("lottery", cur, newVal) {
					break
				}
			}
		}()
	}
	wg.Wait()
	finalVal, _ := smap.Load("lottery")
	fmt.Printf("[sync.Map]    Final lottery value: %v (Expected: 200)\n", finalVal)
}
