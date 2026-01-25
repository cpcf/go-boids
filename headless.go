package main

import (
	"fmt"
	"io"
	"math"
)

type headlessOptions struct {
	frames int
	width  int
	height int
}

type headlessSummary struct {
	frames    int
	width     int
	height    int
	boids     int
	avgSpeed  float64
	minSpeed  float64
	maxSpeed  float64
	centroidX float64
	centroidY float64
}

func runHeadless(cfg config, opts headlessOptions) (headlessSummary, error) {
	if opts.frames < 0 {
		return headlessSummary{}, fmt.Errorf("frames must be non-negative")
	}
	if opts.width <= 0 {
		return headlessSummary{}, fmt.Errorf("width must be positive")
	}
	if opts.height <= 0 {
		return headlessSummary{}, fmt.Errorf("height must be positive")
	}

	sim := newSimulation(cfg)
	sim.Resize(opts.width, opts.height)

	for i := 0; i < opts.frames; i++ {
		sim.Step()
	}

	return summarizeBoids(opts, sim.Boids()), nil
}

func writeHeadlessSummary(w io.Writer, summary headlessSummary) error {
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

func summarizeBoids(opts headlessOptions, boids []boid) headlessSummary {
	boidCount := len(boids)
	if boidCount == 0 {
		return headlessSummary{
			frames:    opts.frames,
			width:     opts.width,
			height:    opts.height,
			boids:     boidCount,
			avgSpeed:  0,
			minSpeed:  0,
			maxSpeed:  0,
			centroidX: 0,
			centroidY: 0,
		}
	}

	totalSpeed := 0.0
	totalX := 0.0
	totalY := 0.0
	minSpeed := math.Inf(1)
	maxSpeed := 0.0
	for _, boid := range boids {
		speed := math.Hypot(boid.vel.x, boid.vel.y)
		totalSpeed += speed
		totalX += boid.pos.x
		totalY += boid.pos.y
		if speed < minSpeed {
			minSpeed = speed
		}
		if speed > maxSpeed {
			maxSpeed = speed
		}
	}

	return headlessSummary{
		frames:    opts.frames,
		width:     opts.width,
		height:    opts.height,
		boids:     boidCount,
		avgSpeed:  totalSpeed / float64(boidCount),
		minSpeed:  minSpeed,
		maxSpeed:  maxSpeed,
		centroidX: totalX / float64(boidCount),
		centroidY: totalY / float64(boidCount),
	}
}
