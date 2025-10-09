# go-boids

A go implementation of the boids simulation described by Craig Reynolds in 1986.

https://dl.acm.org/doi/10.1145/37402.37406

## Output
Runs in the terminal, using the https://github.com/charmbracelet/bubbletea framework to handle terminal output.

Scales the number of boids to the terminal size. Works best on larger terminals. Very large terminal sizes may cause performance issues.

![](https://github.com/cpcf/go-boids/blob/main/go-boids.gif)

## Installation

```bash
go get github.com/cpcf/go-boids
```

## Usage

```bash
go-boids
```

## Parameters
Configurable parameters are set in the `.env` file. Resize window or restart to apply changes.

Defaults and descriptions are as follows:

```bash
FPS               = 120   # Frames per second
BOUNCE            = true  # Bounce off walls
CLAMP_MIN_SPEED   = true  # Sets the minimum speed
TARGET_MIN_SPEED  = 0.05  # Minimum speed of boid if clamped
CELLS_PER_BOID    = 75    # Terminal cells per boid; smaller values mean more boids

RADIUS            = 7.0   # Tiles from boid that affect it
MAX_SPEED         = 1.0   # Maximum speed of boid
ADJUST_RATE       = 0.025 # Rate of adjustment for alignment, cohesion, and separation
ALIGNMENT_RATE    = 1.0   # Muliplier for alignment adjustment
COHESION_RATE     = 1.0   # Muliplier for cohesion adjustment
SEPARATION_RATE   = 1.0   # Muliplier for separation adjustment
```
