package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const defaultCellsPerBoid = 75

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

func loadConfig() config {
	cfg := defaultConfig()
	if err := godotenv.Overload(); err != nil {
		return cfg
	}

	envVars := map[string]interface{}{
		"FPS":              &cfg.fps,
		"BOUNCE":           &cfg.bounce,
		"CLAMP_MIN_SPEED":  &cfg.clampMinSpeed,
		"CELLS_PER_BOID":   &cfg.cellsPerBoid,
		"RADIUS":           &cfg.radius,
		"MAX_SPEED":        &cfg.maxSpeed,
		"ADJUST_RATE":      &cfg.adjustRate,
		"ALIGNMENT_RATE":   &cfg.alignmentRate,
		"COHESION_RATE":    &cfg.cohesionRate,
		"SEPARATION_RATE":  &cfg.separationRate,
		"TARGET_MIN_SPEED": &cfg.targetMinSpeed,
	}

	for key, ptr := range envVars {
		if v := os.Getenv(key); v != "" {
			switch p := ptr.(type) {
			case *int:
				*p, _ = strconv.Atoi(v)
			case *bool:
				*p, _ = strconv.ParseBool(v)
			case *float64:
				*p, _ = strconv.ParseFloat(v, 64)
			}
		}
	}

	if cfg.cellsPerBoid <= 0 {
		cfg.cellsPerBoid = defaultCellsPerBoid
	}

	return cfg
}
