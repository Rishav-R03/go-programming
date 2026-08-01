package exercises

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"strings"
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
