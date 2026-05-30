package main

import (
	"reflect"
	"testing"
)

func TestSimulationStatsOnZeroBoids(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	sim.width = 10
	sim.height = 20

	got := sim.Stats()

	if got.avgSpeed != 0 {
		t.Fatalf("avgSpeed = %v, want 0", got.avgSpeed)
	}
	if got.minSpeed != 0 {
		t.Fatalf("minSpeed = %v, want 0", got.minSpeed)
	}
	if got.maxSpeed != 0 {
		t.Fatalf("maxSpeed = %v, want 0", got.maxSpeed)
	}
	if got.centroidX != 0 {
		t.Fatalf("centroidX = %v, want 0", got.centroidX)
	}
	if got.centroidY != 0 {
		t.Fatalf("centroidY = %v, want 0", got.centroidY)
	}
}

func TestSimulationStatsReportsDimensionsBoidsAndFrames(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	width, height := 40, 24
	sim.Resize(width, height)

	for i := 0; i < 3; i++ {
		sim.Step()
	}

	got := sim.Stats()

	if got.width != width {
		t.Fatalf("width = %v, want %v", got.width, width)
	}
	if got.height != height {
		t.Fatalf("height = %v, want %v", got.height, height)
	}
	if got.boids != len(sim.Boids()) {
		t.Fatalf("boids = %v, want %v", got.boids, len(sim.Boids()))
	}
	if got.frames != 3 {
		t.Fatalf("frames = %v, want %v", got.frames, 3)
	}
}

func TestSimulationStepIncrementsFrames(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	sim.Resize(40, 24)

	sim.frames = 10

	sim.Step()
	if got, want := sim.frames, 11; got != want {
		t.Fatalf("frames = %v, want %v", got, want)
	}

	sim.Step()
	if got, want := sim.frames, 12; got != want {
		t.Fatalf("frames = %v, want %v", got, want)
	}
}

func TestSimulationResizeResetsFrames(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	sim.Resize(40, 24)
	sim.Step()
	sim.Step()

	sim.Resize(30, 12)
	if got, want := sim.frames, 0; got != want {
		t.Fatalf("frames = %v, want %v", got, want)
	}
}

func TestSimulationSetConfigDoesNotResetFrames(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	sim.Resize(40, 24)
	sim.Step()
	before := sim.frames

	beforeBoids := copyBoidsForTest(sim.Boids())
	beforeCfg := sim.cfg

	updatedCfg := cfg
	updatedCfg.radius += 0.5
	sim.SetConfig(updatedCfg)

	if got := sim.frames; got != before {
		t.Fatalf("frames = %v, want %v", got, before)
	}
	if !boidsEqualForTest(beforeBoids, sim.Boids()) {
		t.Fatal("SetConfig should not reset boids")
	}
	if got, want := sim.cfg, updatedCfg; !reflect.DeepEqual(got, want) {
		t.Fatalf("sim.cfg = %#v, want %#v", got, want)
	}
	if reflect.DeepEqual(sim.cfg, beforeCfg) {
		t.Fatal("SetConfig should update sim.cfg")
	}
}

func TestSimulationStatsReadOnly(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	sim.Resize(40, 24)

	beforeBoids := copyBoidsForTest(sim.Boids())
	beforeCfg := sim.cfg
	beforeFrames := sim.frames

	got := sim.Stats()
	if got.frames != beforeFrames {
		t.Fatalf("frames = %v, want %v", got.frames, beforeFrames)
	}

	if !boidsEqualForTest(beforeBoids, sim.Boids()) {
		t.Fatal("Stats() should not mutate boids")
	}
	if !reflect.DeepEqual(sim.cfg, beforeCfg) {
		t.Fatalf("sim.cfg = %#v, want %#v", sim.cfg, beforeCfg)
	}
}

func TestSimulationStatsDoesNotMaterializeBoidView(t *testing.T) {
	cfg := defaultConfig()
	sim := newSimulation(cfg)
	sim.Resize(40, 24)
	sim.Step()

	if !sim.boidsDirty {
		t.Fatal("test setup should leave the compatibility boid view dirty")
	}

	_ = sim.Stats()

	if !sim.boidsDirty {
		t.Fatal("Stats should read SoA state without materializing the compatibility boid view")
	}
}
