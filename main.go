package main

import (
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

func main() {
	updateVars()

	m := model{}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}

var fps = 120
var bouncey = true
var clampMinSpeed = true
var cellsPerBoid = defaultCellsPerBoid

var radius = 7.0
var maxSpeed = 1.0
var adjustRate = 0.025
var alignmentRate = 1.0
var cohesionRate = 1.0
var separationRate = 1.0
var targetMinSpeed = 0.05

const defaultCellsPerBoid = 75

func updateVars() {
	if err := godotenv.Overload(); err != nil {
		// If there's no .env file, we'll just use the defaults
		return
	}
	envVars := map[string]interface{}{
		"FPS":              &fps,
		"BOUNCE":           &bouncey,
		"CLAMP_MIN_SPEED":  &clampMinSpeed,
		"CELLS_PER_BOID":   &cellsPerBoid,
		"RADIUS":           &radius,
		"MAX_SPEED":        &maxSpeed,
		"ADJUST_RATE":      &adjustRate,
		"ALIGNMENT_RATE":   &alignmentRate,
		"COHESION_RATE":    &cohesionRate,
		"SEPARATION_RATE":  &separationRate,
		"TARGET_MIN_SPEED": &targetMinSpeed,
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

	if cellsPerBoid <= 0 {
		cellsPerBoid = defaultCellsPerBoid
	}
}
