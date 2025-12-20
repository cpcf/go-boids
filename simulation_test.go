package main

import "testing"

func TestSimulationStepUsesPreviousFrameState(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 10
	cfg.maxSpeed = 10
	cfg.adjustRate = 1
	cfg.alignmentRate = 1
	cfg.cohesionRate = 0
	cfg.separationRate = 0
	cfg.clampMinSpeed = false

	sim := newSimulation(cfg)
	boids := []boid{
		{
			pos:           Point{x: 10, y: 10},
			vel:           Point{x: 1, y: 0},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		},
		{
			pos:           Point{x: 12, y: 10},
			vel:           Point{x: 3, y: 0},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		},
	}

	sim.Step(boids)

	if got, want := boids[0].vel, (Point{x: 3, y: 0}); got != want {
		t.Fatalf("boid 0 vel = %+v, want %+v", got, want)
	}
	if got, want := boids[1].vel, (Point{x: 1, y: 0}); got != want {
		t.Fatalf("boid 1 vel = %+v, want %+v", got, want)
	}
	if got, want := boids[0].pos, (Point{x: 13, y: 10}); got != want {
		t.Fatalf("boid 0 pos = %+v, want %+v", got, want)
	}
	if got, want := boids[1].pos, (Point{x: 13, y: 10}); got != want {
		t.Fatalf("boid 1 pos = %+v, want %+v", got, want)
	}
}

func TestSimulationStepCountsNeighborsAcrossSpatialGridCellBoundary(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 5
	cfg.maxSpeed = 10
	cfg.adjustRate = 1
	cfg.alignmentRate = 1
	cfg.cohesionRate = 0
	cfg.separationRate = 0
	cfg.clampMinSpeed = false

	sim := newSimulation(cfg)
	boids := padBoidsForSpatialGridPath([]boid{
		{
			pos:           Point{x: 9.9, y: 10},
			vel:           Point{x: 1, y: 0},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		},
		{
			pos:           Point{x: 14.8, y: 10},
			vel:           Point{x: 3, y: 0},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		},
	})

	sim.Step(boids)

	if got, want := boids[0].vel, (Point{x: 3, y: 0}); got != want {
		t.Fatalf("boid 0 vel = %+v, want %+v", got, want)
	}
	if got, want := boids[1].vel, (Point{x: 1, y: 0}); got != want {
		t.Fatalf("boid 1 vel = %+v, want %+v", got, want)
	}
}

func TestSimulationStepDoesNotWrapNeighborSearchAtScreenEdges(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 5
	cfg.maxSpeed = 10
	cfg.adjustRate = 1
	cfg.alignmentRate = 1
	cfg.cohesionRate = 0
	cfg.separationRate = 0
	cfg.clampMinSpeed = false

	sim := newSimulation(cfg)
	boids := []boid{
		{
			pos:           Point{x: 1, y: 10},
			vel:           Point{x: 1, y: 0},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		},
		{
			pos:           Point{x: 29, y: 10},
			vel:           Point{x: 3, y: 0},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		},
	}

	sim.Step(boids)

	if got, want := boids[0].vel, (Point{x: 1, y: 0}); got != want {
		t.Fatalf("boid 0 vel = %+v, want %+v", got, want)
	}
	if got, want := boids[1].vel, (Point{x: 3, y: 0}); got != want {
		t.Fatalf("boid 1 vel = %+v, want %+v", got, want)
	}
}

func padBoidsForSpatialGridPath(boids []boid) []boid {
	for len(boids) < spatialGridBoidThreshold {
		boids = append(boids, boid{
			pos:           Point{x: 25, y: 18},
			vel:           Point{x: 0.1, y: 0.1},
			maxX:          30,
			maxY:          20,
			bounce:        false,
			clampMinSpeed: false,
		})
	}
	return boids
}
