package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func randomInt(min, max int64) int64 {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}

	return min + int64(binary.LittleEndian.Uint64(b[:])%uint64(max-min+1))
}

func main() {
	// Random duration from 20 through 30 seconds.
	duration := time.Duration(randomInt(20, 30)) * time.Second

	// Allocate up to 1 GiB, but keep the actual allocation bounded by
	// available memory and avoid making the host unusable.
	const maxMemory = 1 << 30
	memory := make([]byte, maxMemory)

	// Touch pages so the allocation is backed by physical memory as needed.
	for i := 0; i < len(memory); i += 4096 {
		memory[i] = byte(i)
	}

	fmt.Printf("Simulating activity for %v\n", duration)

	deadline := time.Now().Add(duration)
	var wg sync.WaitGroup

	for time.Now().Before(deadline) {
		
		// Randomly vary the number of CPU workers for this tick.
		workers := int(randomInt(1, int64(runtime.NumCPU())))
		tick := time.Duration(randomInt(100, 500)) * time.Millisecond


		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()

				end := time.Now().Add(tick)
				var value uint64

				for time.Now().Before(end) {
					value ^= value<<13 + value>>7 + 0x9e3779b97f4a7c15
				}
			}()
		}

		// Randomly touch part of the allocated memory during each tick.
		bytesToTouch := int(randomInt(1<<20, maxMemory))
		for i := 0; i < bytesToTouch; i += 4096 {
			memory[i]++
		}
		fmt.Printf("Ticking with %d workers and %d bytes\n", workers, bytesToTouch)

		wg.Wait()
		time.Sleep(time.Duration(randomInt(100, 500)) * time.Millisecond)
	}

	// Release memory and exit cleanly.
	memory = nil
	runtime.GC()

	fmt.Println("Simulation complete.")
}