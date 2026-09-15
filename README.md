# System Activity Simulator

A small Go program that simulates variable CPU and memory activity for a
random duration between 20 and 30 seconds.

## Features

- Runs for a randomly selected duration from 20–30 seconds
- Uses a randomly varying number of CPU workers
- Allocates up to 1 GiB of memory
- Randomly varies memory activity during each tick
- Releases resources and exits cleanly

## Requirements

- Go 1.20 or later
- A system with sufficient available memory

## Running the application

From the project directory, run:

```bash
go run main.go
```

To build and run a standalone executable:

```bash
go build -o activity-simulator .
./activity-simulator
```

## Example output

```text
Simulating activity for 24s
Simulation complete.
```

## Warning

This application can consume substantial CPU and memory. Do not run it on
production systems or machines with limited resources unless that behavior is
intentional.