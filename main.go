package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const (
	maxMemory = 1 << 30
	pageSize  = 4096
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	duration := randomDuration(rng, 10*time.Second, 20*time.Second)
	fmt.Printf(
		"Starting system activity simulation for %v using up to %d CPU workers\n",
		duration,
		runtime.NumCPU()*2,
	)

	memory := make([]byte, maxMemory)
	fmt.Printf("Allocated %d MiB of simulated RAM\n", len(memory)/(1<<20))

	fmt.Println("Touching memory pages to commit the allocation...")
	touchAllPages(memory)
	fmt.Println("Initial memory commitment complete")

	// Give separate instances different startup phases.
	startupDelay := randomDuration(rng, 0, 1500*time.Millisecond)
	if startupDelay > 0 {
		fmt.Printf(
			"Delaying activity start by %v to vary instance timing\n",
			startupDelay,
		)
		time.Sleep(startupDelay)
	}

	deadline := time.Now().Add(duration)

	// Keep the worker count bounded. Creating and destroying many goroutines
	// every tick tends to produce a synchronized sawtooth workload.
	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers < 2 {
		maxWorkers = 2
	}
	fmt.Printf("Starting %d independent CPU activity workers\n", maxWorkers)
	var stop atomic.Bool
	var wg sync.WaitGroup

	// Each worker independently alternates between busy and idle periods.
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			workerRNG := rand.New(rand.NewSource(
				time.Now().UnixNano() + int64(workerID)*7919,
			))

			var value uint64

			for !stop.Load() {
				// The worker does not always participate. This makes CPU
				// utilization less synchronized between workers.
				if workerRNG.Intn(100) < 70 {
					busyFor := randomDuration(
						workerRNG,
						40*time.Millisecond,
						700*time.Millisecond,
					)
					fmt.Printf(
						"CPU worker %d active for approximately %v\n",
						workerID,
						busyFor,
					)

					end := time.Now().Add(busyFor)
					for time.Now().Before(end) {
						value ^= value<<13 + value>>7 +
							0x9e3779b97f4a7c15
					}
				} else {
					idleFor := randomDuration(
						workerRNG,
						20*time.Millisecond,
						500*time.Millisecond,
					)

					fmt.Printf(
						"CPU worker %d temporarily idle for approximately %v\n",
						workerID,
						idleFor,
					)

					time.Sleep(idleFor)
				}
			}

			// Prevent the compiler from eliminating the calculation.
			_ = value
		}(i)
	}

	// Memory activity runs independently of the CPU workers.
	for time.Now().Before(deadline) {
		fmt.Println("Selecting random memory regions to touch")
		touchRandomMemory(rng, memory)

		// Irregular intervals prevent all instances from producing identical
		// chart boundaries.
		pause := randomDuration(
			rng,
			80*time.Millisecond,
			900*time.Millisecond,
		)
		fmt.Printf(
			"Memory activity pass complete; waiting approximately %v\n",
			pause,
		)
		time.Sleep(pause)
	}
	fmt.Println("Simulation duration reached; stopping CPU workers")
	stop.Store(true)
	wg.Wait()

	// Release memory and exit cleanly.
	fmt.Println("Releasing simulated RAM and requesting garbage collection")
	memory = nil
	runtime.GC()

	fmt.Println("Simulation complete.")
}

func randomDuration(rng *rand.Rand, min, max time.Duration) time.Duration {
	if max <= min {
		return min
	}

	return min + time.Duration(rng.Int63n(int64(max-min)+1))
}

func touchAllPages(memory []byte) {
	for offset := 0; offset < len(memory); offset += pageSize {
		memory[offset] = byte(offset / pageSize)
	}
}

func touchRandomMemory(rng *rand.Rand, memory []byte) {
	if len(memory) == 0 {
		fmt.Println("Skipping memory activity because no memory is allocated")
		return
	}

	// Vary the amount of memory touched substantially.
	minPages := 256                    // 1 MiB
	maxPages := len(memory) / pageSize // Up to the full allocation
	pageCount := minPages

	if maxPages > minPages {
		pageCount += rng.Intn(maxPages - minPages + 1)
	}

	// Touch several independent regions rather than one linear range.
	regions := 1 + rng.Intn(8)

	fmt.Printf(
		"Touching approximately %d MiB across %d random memory regions\n",
		(pageCount*pageSize)/(1<<20),
		regions,
	)

	for region := 0; region < regions; region++ {
		maxStartPage := maxPages - pageCount
		startPage := 0

		if maxStartPage > 0 {
			startPage = rng.Intn(maxStartPage + 1)
		}

		for page := 0; page < pageCount/regions; page++ {
			offset := (startPage + page) * pageSize
			if offset >= len(memory) {
				break
			}

			memory[offset]++
		}
	}
}
