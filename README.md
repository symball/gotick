# System Activity Simulator

A small Go program that simulates variable CPU and RAM activity for a configurable period. It is useful for testing monitoring dashboards, workload visualizations, runner lifecycle events, and alerting systems.

The simulator creates independent CPU workers that alternate between busy and idle periods while gradually changing its memory target and touching the committed memory.

## Features

- Runs for a random duration between 10 and 20 seconds
- Gradually varies its memory target between 16 MiB and 2 GiB
- Grows and shrinks the memory allocation instead of allocating the maximum immediately
- Commits newly allocated memory by touching its pages
- Randomly selects a new memory target during the simulation
- Uses multiple independent CPU activity workers
- Randomly varies CPU busy and idle periods
- Touches the current memory allocation during each activity pass
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
Memory target will vary between 16 MiB and 2048 MiB
Delaying activity start by 734ms to vary instance timing
Starting 16 independent CPU activity workers
CPU worker 3 active for approximately 412ms
CPU worker 7 temporarily idle for approximately 183ms
Increasing memory target from 16 MiB to 384 MiB
Touching the current 384 MiB memory target
Memory activity pass complete; waiting approximately 527ms
Simulation duration reached; stopping CPU workers
All CPU workers stopped
Releasing 128 MiB of simulated RAM and requesting garbage collection
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

- Uses a randomly selected memory target between 16 MiB and 2 GiB
- Changes the memory target during the simulation
- Allocates memory gradually as the target increases
- Releases memory when the target decreases
- Uses up to twice the number of available CPU cores as workers
- Performs CPU-intensive calculations
  - Touches memory pages whenever memory is allocated or an activity pass runs

Running several instances can consume significant system resources, particularly if multiple instances select large memory targets at the same time. Start with one instance and monitor the host before increasing the number of concurrent processes.

## Stopping the Program

The program normally exits after its randomly selected duration.

To stop it manually, press:

```text
Ctrl+C
```

Because the program is intentionally CPU-intensive, use caution when running it on production systems or shared machines.

## How It Works

1. A random simulation duration is selected.
2. A memory target is selected in 16 MiB increments, up to 2 GiB.
3. The allocation grows or shrinks toward the selected target.
4. Newly allocated pages are touched so the memory is backed by physical memory.
5. CPU workers independently alternate between busy and idle periods.
6. The memory target is changed repeatedly during the simulation.
7. The current allocation is touched during each memory activity pass.
8. The program logs CPU and memory activity as it runs.
9. Workers stop, memory is released, and garbage collection is requested.

## Limitations

This program is intended for testing and demonstration purposes. It does not represent a realistic application workload and should not be used as a benchmark.

The memory limits, memory step size, and CPU worker count are currently defined in the source code. Adjust them carefully if running on a system with limited resources. A 2 GiB target may cause memory pressure when several instances run concurrently.