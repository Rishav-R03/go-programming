package main

import "hashmaps/exercises"

/*
Note on Sys vs Alloc: Notice Alloc drops back to 0 MB, but Sys stays around 1091 MB. Go holds onto this memory
from the OS for a while (using MADV_DONTNEED or MADV_FREE) so it can reuse it quickly for future allocations rather
than repeatedly asking the OS kernel for RAM.
*/
func main() {
	exercises.ExerciseOneFixed()
}
