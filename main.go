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
	maxMemory     = 2 << 30
	minMemoryStep = 16 << 20
	pageSize      = 4096
)

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	duration := randomDuration(rng, 10*time.Second, 20*time.Second)
	fmt.Printf(
		"Starting system activity simulation for %v using up to %d CPU workers\n",
		duration,
		runtime.NumCPU()*2,
	)

	memory := make([]byte, 0, maxMemory)
	fmt.Printf(
		"Memory target will vary between %d MiB and %d MiB\n",
		minMemoryStep/(1<<20),
		maxMemory/(1<<20),
	)

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
		targetMemory := randomMemoryTarget(rng)
		memory = adjustMemoryTarget(memory, targetMemory)

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
	fmt.Printf(
		"Releasing %d MiB of simulated RAM and requesting garbage collection\n",
		len(memory)/(1<<20),
	)
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

func randomMemoryTarget(rng *rand.Rand) int {
	steps := maxMemory / minMemoryStep
	return minMemoryStep * (1 + rng.Intn(steps))
}

func adjustMemoryTarget(memory []byte, target int) []byte {
	current := len(memory)

	if current == target {
		fmt.Printf("Memory target remains at %d MiB\n", target/(1<<20))
		touchRandomMemory(memory)
		return memory
	}

	if current < target {
		growth := target - current
		fmt.Printf(
			"Increasing memory target from %d MiB to %d MiB\n",
			current/(1<<20),
			target/(1<<20),
		)

		memory = append(memory, make([]byte, growth)...)
		touchRange(memory, current, len(memory))
		return memory
	}

	fmt.Printf(
		"Decreasing memory target from %d MiB to %d MiB\n",
		current/(1<<20),
		target/(1<<20),
	)

	memory = memory[:target]
	runtime.GC()
	return memory
}

func touchRandomMemory(memory []byte) {
	if len(memory) == 0 {
		fmt.Println("Skipping memory activity because no memory is allocated")
		return
	}

	fmt.Printf(
		"Touching the current %d MiB memory target\n",
		len(memory)/(1<<20),
	)
	touchRange(memory, 0, len(memory))
}

func touchRange(memory []byte, start, end int) {
	if start < 0 {
		start = 0
	}
	if end > len(memory) {
		end = len(memory)
	}

	for offset := start; offset < end; offset += pageSize {
		memory[offset]++
	}
}
