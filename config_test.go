package main

import "testing"

func TestDefaultConfigReturnsExpectedValues(t *testing.T) {
	got := defaultConfig()

	want := config{
		fps:            120,
		bounce:         true,
		clampMinSpeed:  true,
		cellsPerBoid:   defaultCellsPerBoid,
		radius:         7.0,
		maxSpeed:       1.0,
		adjustRate:     0.025,
		alignmentRate:  1.0,
		cohesionRate:   1.0,
		separationRate: 1.0,
		targetMinSpeed: 0.05,
	}

	if got != want {
		t.Fatalf("defaultConfig() = %+v, want %+v", got, want)
	}
}
