# System Activity Simulator

A small Go program that simulates variable CPU and RAM activity for a configurable period. It is useful for testing monitoring dashboards, workload visualizations, runner lifecycle events, and alerting systems.

The simulator creates independent CPU workers that alternate between busy and idle periods while randomly touching different regions of an allocated memory buffer.

## Features

- Runs for a random duration between 10 and 20 seconds
- Allocates and commits up to 1 GiB of memory
- Uses multiple independent CPU activity workers
- Randomly varies CPU busy and idle periods
- Touches different memory regions during each activity pass
- Adds startup timing variation between instances
- Logs CPU and memory activity to standard output
- Cleans up memory and exits normally when complete

## Requirements

- Go 1.20 or newer
- Sufficient available memory for the allocation
- Permission to run CPU-intensive workloads

## Running

Save the program as `main.go`, then run:

```bash
go run .
```

Alternatively, build and execute it:

```bash
go build -o activity-simulator .
./activity-simulator
```

## Example Output

```text
Starting system activity simulation for 16.8s using up to 16 CPU workers
Allocated 1024 MiB of simulated RAM
Touching memory pages to commit the allocation...
Initial memory commitment complete
Delaying activity start by 734ms to vary instance timing
Starting 16 independent CPU activity workers
CPU worker 3 active for approximately 412ms
CPU worker 7 temporarily idle for approximately 183ms
Selecting random memory regions to touch
Touching approximately 386 MiB across 5 random memory regions
Memory activity pass complete; waiting approximately 527ms
Simulation duration reached; stopping CPU workers
All CPU workers stopped
Releasing simulated RAM and requesting garbage collection
Simulation complete.
```

The exact output and timing vary on every run.

## Running Multiple Instances

Multiple instances can be started independently:

```bash
./activity-simulator &
./activity-simulator &
./activity-simulator &
wait
```

Each instance uses a different random seed and startup delay, so their CPU and memory activity should not be perfectly synchronized.

## Resource Usage

By default, the program:

- Allocates approximately 1 GiB of memory
- Uses up to twice the number of available CPU cores as workers
- Performs CPU-intensive calculations
- Randomly touches memory pages throughout the run

Running several instances can consume significant system resources. Start with one instance and monitor the host before increasing the number of concurrent processes.

## Stopping the Program

The program normally exits after its randomly selected duration.

To stop it manually, press:

```text
Ctrl+C
```

Because the program is intentionally CPU-intensive, use caution when running it on production systems or shared machines.

## How It Works

1. A random simulation duration is selected.
2. A 1 GiB memory buffer is allocated.
3. Each memory page is touched so the allocation is backed by physical memory.
4. CPU workers independently alternate between busy and idle periods.
5. Random memory regions are touched at irregular intervals.
6. The program logs activity as it runs.
7. Workers stop, memory is released, and garbage collection is requested.

## Limitations

This program is intended for testing and demonstration purposes. It does not represent a realistic application workload and should not be used as a benchmark.

The memory allocation and CPU worker count are currently defined in the source code. Adjust them carefully if running on a system with limited resources.