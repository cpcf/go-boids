package main

import (
	"fmt"
	"io"
)

type headlessOptions struct {
	frames int
	width  int
	height int
}

func runHeadless(cfg config, opts headlessOptions) (simulationStats, error) {
	if opts.frames < 0 {
		return simulationStats{}, fmt.Errorf("frames must be non-negative")
	}
	if opts.width <= 0 {
		return simulationStats{}, fmt.Errorf("width must be positive")
	}
	if opts.height <= 0 {
		return simulationStats{}, fmt.Errorf("height must be positive")
	}

	sim := newSimulation(cfg)
	sim.Resize(opts.width, opts.height)

	for i := 0; i < opts.frames; i++ {
		sim.Step()
	}

	return sim.Stats(), nil
}

func writeHeadlessSummary(w io.Writer, summary simulationStats) error {
	_, err := fmt.Fprintf(
		w,
		"frames=%d width=%d height=%d boids=%d avg_speed=%.3f min_speed=%.3f max_speed=%.3f centroid_x=%.3f centroid_y=%.3f\n",
		summary.frames,
		summary.width,
		summary.height,
		summary.boids,
		summary.avgSpeed,
		summary.minSpeed,
		summary.maxSpeed,
		summary.centroidX,
		summary.centroidY,
	)
	return err
}
