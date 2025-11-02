package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
	cfg := loadConfig()
	fps = cfg.fps
	bouncey = cfg.bounce
	clampMinSpeed = cfg.clampMinSpeed
	cellsPerBoid = cfg.cellsPerBoid
	radius = cfg.radius
	maxSpeed = cfg.maxSpeed
	adjustRate = cfg.adjustRate
	alignmentRate = cfg.alignmentRate
	cohesionRate = cfg.cohesionRate
	separationRate = cfg.separationRate
	targetMinSpeed = cfg.targetMinSpeed
}
