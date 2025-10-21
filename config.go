package main

type config struct {
	fps            int
	bounce         bool
	clampMinSpeed  bool
	cellsPerBoid   int
	radius         float64
	maxSpeed       float64
	adjustRate     float64
	alignmentRate  float64
	cohesionRate   float64
	separationRate float64
	targetMinSpeed float64
}

func defaultConfig() config {
	return config{
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
}
