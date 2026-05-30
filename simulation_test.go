package main

import (
	"math"
	"testing"
)

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
	sim.Resize(30, 20)
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

func TestSimulationSetConfigUpdatesConfigAndGridCellSizeWithoutReset(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(123)
	sim := newSimulation(cfg)
	sim.Resize(20, 10)

	before := copyBoidsForTest(sim.Boids())

	updatedCfg := cfg
	updatedCfg.radius = cfg.radius + 2.5
	updatedCfg.maxSpeed = cfg.maxSpeed + 0.3
	sim.SetConfig(updatedCfg)

	if sim.cfg != updatedCfg {
		t.Fatalf("sim.cfg = %+v, want %+v", sim.cfg, updatedCfg)
	}
	if got, want := sim.nearbyGrid.cellSize, updatedCfg.radius; got != want {
		t.Fatalf("nearby grid cell size = %.6f, want %.6f", got, want)
	}
	if !boidsEqualForTest(before, sim.Boids()) {
		t.Fatal("SetConfig should not reset boids")
	}
}

func TestSimulationSetConfigFallsBackInvalidGridCellSize(t *testing.T) {
	invalidRadii := []float64{0, -1, math.NaN(), math.Inf(1), math.Inf(-1)}

	for _, radius := range invalidRadii {
		cfg := defaultConfig()
		cfg.seed = uint64Ptr(123)
		sim := newSimulation(cfg)
		sim.Resize(20, 10)

		updatedCfg := cfg
		updatedCfg.radius = radius
		sim.SetConfig(updatedCfg)

		if got, want := sim.nearbyGrid.cellSize, float64(defaultSpatialGridCellSize); got != want {
			t.Fatalf("radius=%v nearby grid cell size = %.6f, want %.6f", radius, got, want)
		}
	}
}

func TestSimulationStepSwapsBoidBuffers(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 0
	cfg.maxSpeed = 10
	cfg.adjustRate = 1
	cfg.alignmentRate = 0
	cfg.cohesionRate = 0
	cfg.separationRate = 0
	cfg.clampMinSpeed = false

	sim := newSimulation(cfg)
	sim.boids = []boid{
		{
			pos:           Point{x: 1, y: 1},
			vel:           Point{x: 1, y: 0},
			maxX:          5,
			maxY:          5,
			bounce:        false,
			clampMinSpeed: false,
		},
		{
			pos:           Point{x: 2, y: 2},
			vel:           Point{x: -1, y: 0},
			maxX:          5,
			maxY:          5,
			bounce:        false,
			clampMinSpeed: false,
		},
	}

	initialCurrentBoid0 := &sim.boids[0]
	initialCurrentBoid1 := &sim.boids[1]

	sim.Step()
	afterFirst := sim.Boids()
	if len(sim.nextBoids) != len(sim.boids) {
		t.Fatalf("len(sim.nextBoids) = %d, want %d", len(sim.nextBoids), len(sim.boids))
	}
	if &sim.nextBoids[0] != initialCurrentBoid0 || &sim.nextBoids[1] != initialCurrentBoid1 {
		t.Fatal("after first step, next buffer should reuse previous frame backing array")
	}

	firstCurrent := &afterFirst[0]
	firstCurrent1 := &afterFirst[1]
	sim.Step()
	afterSecond := sim.Boids()

	if &afterSecond[0] != initialCurrentBoid0 || &afterSecond[1] != initialCurrentBoid1 {
		t.Fatal("after second step, current frame should have swapped back to the initial backing array")
	}
	if &sim.nextBoids[0] != firstCurrent || &sim.nextBoids[1] != firstCurrent1 {
		t.Fatal("after second step, next buffer should be the buffer written by the previous step")
	}

	if afterSecond[0] == afterFirst[0] {
		t.Fatal("positions should advance across steps")
	}
}
