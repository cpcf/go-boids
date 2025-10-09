package main

import "testing"

func TestUpdateBoidsUsesPreviousFrameState(t *testing.T) {
	preserveSimGlobals(t)
	radius = 10
	maxSpeed = 10
	adjustRate = 1
	alignmentRate = 1
	cohesionRate = 0
	separationRate = 0
	clampMinSpeed = false

	var m model
	m.cells.init(30, 20)
	m.boids = []boid{
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

	m.updateBoids()

	if got, want := m.boids[0].vel, (Point{x: 3, y: 0}); got != want {
		t.Fatalf("boid 0 vel = %+v, want %+v", got, want)
	}
	if got, want := m.boids[1].vel, (Point{x: 1, y: 0}); got != want {
		t.Fatalf("boid 1 vel = %+v, want %+v", got, want)
	}
	if got, want := m.boids[0].pos, (Point{x: 13, y: 10}); got != want {
		t.Fatalf("boid 0 pos = %+v, want %+v", got, want)
	}
	if got, want := m.boids[1].pos, (Point{x: 13, y: 10}); got != want {
		t.Fatalf("boid 1 pos = %+v, want %+v", got, want)
	}
}

func TestUpdateBoidsCountsNeighborsAcrossSpatialGridCellBoundary(t *testing.T) {
	preserveSimGlobals(t)
	radius = 5
	maxSpeed = 10
	adjustRate = 1
	alignmentRate = 1
	cohesionRate = 0
	separationRate = 0
	clampMinSpeed = false

	var m model
	m.cells.init(30, 20)
	m.boids = []boid{
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
	}
	m.boids = padBoidsForSpatialGridPath(m.boids)

	m.updateBoids()

	if got, want := m.boids[0].vel, (Point{x: 3, y: 0}); got != want {
		t.Fatalf("boid 0 vel = %+v, want %+v", got, want)
	}
	if got, want := m.boids[1].vel, (Point{x: 1, y: 0}); got != want {
		t.Fatalf("boid 1 vel = %+v, want %+v", got, want)
	}
}

func TestUpdateBoidsDoesNotWrapNeighborSearchAtScreenEdges(t *testing.T) {
	preserveSimGlobals(t)
	radius = 5
	maxSpeed = 10
	adjustRate = 1
	alignmentRate = 1
	cohesionRate = 0
	separationRate = 0
	clampMinSpeed = false

	var m model
	m.cells.init(30, 20)
	m.boids = []boid{
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

	m.updateBoids()

	if got, want := m.boids[0].vel, (Point{x: 1, y: 0}); got != want {
		t.Fatalf("boid 0 vel = %+v, want %+v", got, want)
	}
	if got, want := m.boids[1].vel, (Point{x: 3, y: 0}); got != want {
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
