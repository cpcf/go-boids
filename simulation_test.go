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
	sim.boids = []boid{
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

	sim.Step()
	boids := sim.Boids()

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
	sim.boids = padBoidsForSpatialGridPath([]boid{
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

	sim.Step()
	boids := sim.Boids()

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
	sim.boids = []boid{
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

	sim.Step()
	boids := sim.Boids()

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

func TestSimulationResizeConfiguresDimensionsAndBoids(t *testing.T) {
	cfg := defaultConfig()
	cfg.cellsPerBoid = 100

	sim := newSimulation(cfg)
	width, height := 120, 40
	sim.Resize(width, height)

	if got, want := sim.width, width; got != want {
		t.Fatalf("sim.width = %v, want %v", got, want)
	}
	if got, want := sim.height, height; got != want {
		t.Fatalf("sim.height = %v, want %v", got, want)
	}

	wantCount := width*height/effectiveCellsPerBoid(cfg) + 1
	if got := len(sim.Boids()); got != wantCount {
		t.Fatalf("len(sim.Boids()) = %v, want %v", got, wantCount)
	}

	for _, b := range sim.Boids() {
		if got, want := b.maxX, float64(width); got != want {
			t.Fatalf("boid maxX = %v, want %v", got, want)
		}
		if got, want := b.maxY, float64(height); got != want {
			t.Fatalf("boid maxY = %v, want %v", got, want)
		}
	}
}
