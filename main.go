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
	maxMemory = 2 << 30 // 2 GiB
	pageSize  = 4096

	// Memory is changed in chunks rather than allocated all at once.
	minMemoryStep = 8 << 20   // 8 MiB
	maxMemoryStep = 128 << 20 // 128 MiB

	// Keep the process from immediately jumping to the maximum allocation.
	initialMemoryMin = 8 << 20  // 8 MiB
	initialMemoryMax = 64 << 20 // 64 MiB
)

type memoryBlock struct {
	data []byte
}

func main() {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	duration := randomDuration(
		rng,
		10*time.Second,
		20*time.Second,
	)

	maxWorkers := runtime.NumCPU() * 2
	if maxWorkers < 2 {
		maxWorkers = 2
	}

	fmt.Printf(
		"Starting system activity simulation for %v with up to %d CPU workers\n",
		duration,
		maxWorkers,
	)

	// Different instances start with different memory levels. This prevents
	// three processes launched together from producing the same first graph.
	initialTarget := randomAlignedSize(
		rng,
		initialMemoryMin,
		initialMemoryMax,
		minMemoryStep,
	)

	memory := make([]byte, 0, maxMemory)

	fmt.Printf(
		"Starting with a random memory target of %s\n",
		formatBytes(initialTarget),
	)

	memory = growMemory(memory, initialTarget, rng)

	// Add startup jitter so simultaneously launched processes do not share
	// the same timeline.
	startupDelay := randomDuration(
		rng,
		0,
		2500*time.Millisecond,
	)

	if startupDelay > 0 {
		fmt.Printf(
			"Waiting %v before starting activity\n",
			startupDelay,
		)
		time.Sleep(startupDelay)
	}

	deadline := time.Now().Add(duration)

	var stop atomic.Bool
	var wg sync.WaitGroup

	fmt.Printf("Starting %d independent CPU workers\n", maxWorkers)

	for workerID := 0; workerID < maxWorkers; workerID++ {
		wg.Add(1)

		go cpuWorker(
			workerID,
			&stop,
			&wg,
			rng.Int63(),
		)
	}

	// Memory activity has its own irregular schedule. It does not use the
	// same timing as the CPU workers.
	runMemoryActivity(
		rng,
		&memory,
		deadline,
	)

	fmt.Println("Simulation duration reached")
	fmt.Println("Stopping CPU workers")

	stop.Store(true)
	wg.Wait()

	fmt.Println("All CPU workers stopped")

	fmt.Printf(
		"Releasing %s of simulated RAM\n",
		formatBytes(len(memory)),
	)

	memory = nil
	runtime.GC()

	fmt.Println("Simulation complete")
}

func cpuWorker(
	workerID int,
	stop *atomic.Bool,
	wg *sync.WaitGroup,
	seed int64,
) {
	defer wg.Done()

	rng := rand.New(rand.NewSource(seed))

	// Each worker gets a different initial phase. This avoids all workers
	// becoming busy immediately when the process starts.
	initialDelay := randomDuration(
		rng,
		0,
		1800*time.Millisecond,
	)

	time.Sleep(initialDelay)

	var value uint64

	for !stop.Load() {
		// Workers do not all use the CPU at the same time. The probability
		// and duration are independently randomized per worker.
		if rng.Intn(100) < 72 {
			busyDuration := randomDuration(
				rng,
				25*time.Millisecond,
				900*time.Millisecond,
			)

			end := time.Now().Add(busyDuration)

			for time.Now().Before(end) {
				value ^= value << 13
				value ^= value >> 7
				value += 0x9e3779b97f4a7c15
			}
		} else {
			idleDuration := randomDuration(
				rng,
				30*time.Millisecond,
				1100*time.Millisecond,
			)

			time.Sleep(idleDuration)
		}

		// Occasionally add a longer pause. This creates more natural gaps
		// between bursts and makes separate processes diverge further.
		if rng.Intn(100) < 12 {
			time.Sleep(randomDuration(
				rng,
				200*time.Millisecond,
				1800*time.Millisecond,
			))
		}
	}

	// Ensure the calculation remains observable to the compiler.
	_ = value

	fmt.Printf("CPU worker %d stopped\n", workerID)
}

func runMemoryActivity(
	rng *rand.Rand,
	memory *[]byte,
	deadline time.Time,
) {
	nextLog := time.Now()

	for time.Now().Before(deadline) {
		currentSize := len(*memory)

		// Choose a new target based partly on the current allocation. This
		// produces a wandering memory curve rather than random vertical jumps.
		targetSize := chooseMemoryTarget(rng, currentSize)

		if targetSize > currentSize {
			*memory = growMemory(*memory, targetSize, rng)
		} else if targetSize < currentSize {
			*memory = shrinkMemory(*memory, targetSize, rng)
		} else {
			touchRandomMemory(*memory, rng)
		}

		// Periodically touch random regions even when the target changes only
		// slightly. This keeps memory activity varied without always changing
		// the resident size.
		if rng.Intn(100) < 70 {
			touchRandomMemory(*memory, rng)
		}

		// Avoid excessive log output while retaining useful state messages.
		if time.Now().After(nextLog) {
			fmt.Printf(
				"Memory activity at %s; target is %s\n",
				time.Now().Format("15:04:05.000"),
				formatBytes(len(*memory)),
			)

			nextLog = time.Now().Add(
				randomDuration(
					rng,
					500*time.Millisecond,
					1500*time.Millisecond,
				),
			)
		}

		// Different instances use different wait periods, so their memory
		// changes do not line up on the same chart samples.
		time.Sleep(randomDuration(
			rng,
			100*time.Millisecond,
			1400*time.Millisecond,
		))
	}
}

func chooseMemoryTarget(rng *rand.Rand, current int) int {
	// Occasionally make a larger move so the chart has visible variation.
	if rng.Intn(100) < 18 {
		return randomAlignedSize(
			rng,
			minMemoryStep,
			maxMemory,
			minMemoryStep,
		)
	}

	// Most changes are relative to the current value. This creates a
	// wandering, non-linear memory graph.
	direction := -1
	if rng.Intn(100) < 56 {
		direction = 1
	}

	step := randomAlignedSize(
		rng,
		minMemoryStep,
		maxMemoryStep,
		minMemoryStep,
	)

	target := current + direction*step

	// Keep the target inside the valid range.
	if target < minMemoryStep {
		target = minMemoryStep
	}
	if target > maxMemory {
		target = maxMemory
	}

	return alignDown(target, minMemoryStep)
}

func growMemory(
	memory []byte,
	target int,
	rng *rand.Rand,
) []byte {
	if target <= len(memory) {
		return memory
	}

	fmt.Printf(
		"Increasing memory from %s to %s\n",
		formatBytes(len(memory)),
		formatBytes(target),
	)

	for len(memory) < target {
		remaining := target - len(memory)
		step := randomAlignedSize(
			rng,
			minMemoryStep,
			maxMemoryStep,
			pageSize,
		)

		if step > remaining {
			step = remaining
		}

		oldLength := len(memory)
		memory = append(memory, make([]byte, step)...)

		// Touch only the newly allocated range. This commits the pages
		// gradually instead of faulting in the entire target at once.
		touchRange(memory, oldLength, len(memory))
	}

	return memory
}

func shrinkMemory(
	memory []byte,
	target int,
	rng *rand.Rand,
) []byte {
	if target >= len(memory) {
		return memory
	}

	fmt.Printf(
		"Decreasing memory from %s to %s\n",
		formatBytes(len(memory)),
		formatBytes(target),
	)

	for len(memory) > target {
		current := len(memory)
		step := randomAlignedSize(
			rng,
			minMemoryStep,
			maxMemoryStep,
			pageSize,
		)

		next := current - step
		if next < target {
			next = target
		}

		memory = memory[:next]

		// Give the runtime an opportunity to release unused backing memory.
		// The periodic collection is intentionally randomized so multiple
		// processes do not all collect at the same time.
		if rng.Intn(100) < 35 {
			runtime.GC()
		}
	}

	return memory
}

func touchRandomMemory(memory []byte, rng *rand.Rand) {
	if len(memory) == 0 {
		return
	}

	pageCount := len(memory) / pageSize
	if pageCount == 0 {
		return
	}

	// Touch a variable number of pages in several separated regions.
	regions := 1 + rng.Intn(8)
	pagesPerRegion := 64 + rng.Intn(maxInt(64, pageCount/8))

	fmt.Printf(
		"Touching %s across %d random memory regions\n",
		formatBytes(minInt(
			len(memory),
			regions*pagesPerRegion*pageSize,
		)),
		regions,
	)

	for region := 0; region < regions; region++ {
		startPage := rng.Intn(pageCount)
		pages := pagesPerRegion

		if startPage+pages > pageCount {
			pages = pageCount - startPage
		}

		start := startPage * pageSize
		end := start + pages*pageSize

		touchRange(memory, start, end)
	}
}

func touchRange(memory []byte, start, end int) {
	if start < 0 {
		start = 0
	}
	if end > len(memory) {
		end = len(memory)
	}
	if start >= end {
		return
	}

	for offset := start; offset < end; offset += pageSize {
		memory[offset]++
	}
}

func randomAlignedSize(
	rng *rand.Rand,
	minimum int,
	maximum int,
	alignment int,
) int {
	if maximum <= minimum {
		return alignDown(minimum, alignment)
	}

	minimum = alignUp(minimum, alignment)
	maximum = alignDown(maximum, alignment)

	if maximum <= minimum {
		return minimum
	}

	count := ((maximum - minimum) / alignment) + 1
	return minimum + rng.Intn(count)*alignment
}

func randomDuration(
	rng *rand.Rand,
	minimum time.Duration,
	maximum time.Duration,
) time.Duration {
	if maximum <= minimum {
		return minimum
	}

	return minimum + time.Duration(
		rng.Int63n(int64(maximum-minimum)+1),
	)
}

func alignUp(value, alignment int) int {
	if alignment <= 1 {
		return value
	}

	remainder := value % alignment
	if remainder == 0 {
		return value
	}

	return value + alignment - remainder
}

func alignDown(value, alignment int) int {
	if alignment <= 1 {
		return value
	}

	return value - value%alignment
}

func formatBytes(value int) string {
	switch {
	case value >= 1<<30:
		return fmt.Sprintf("%.2f GiB", float64(value)/(1<<30))
	case value >= 1<<20:
		return fmt.Sprintf("%.0f MiB", float64(value)/(1<<20))
	case value >= 1<<10:
		return fmt.Sprintf("%.0f KiB", float64(value)/(1<<10))
	default:
		return fmt.Sprintf("%d B", value)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}

	return b
}
