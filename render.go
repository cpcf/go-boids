package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type frameMsg struct{}

const spatialGridBoidThreshold = 128

func animate(cfg config) tea.Cmd {
	if cfg == (config{}) {
		cfg = defaultConfig()
	}

	return tea.Tick(time.Second/time.Duration(cfg.fps), func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	cfg              config
	cells            cellbuffer
	boids            []boid
	previousBoids    []boid
	nearbyGrid       spatialGrid
	candidateIndexes []int
}

func (m model) Init() tea.Cmd {
	return animate(m.cfg)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.cfg = updateVars()
		m.cells.init(msg.Width, msg.Height)
		m.boids = initBoidsOnScreenSize(m.cfg, msg.Width, msg.Height)
		return m, nil
	case frameMsg:
		if !m.cells.ready() {
			return m, nil
		}

		m.cells.wipe()
		m.updateBoids()
		return m, animate(m.cfg)
	default:
		return m, nil
	}
}

func (m model) View() string {
	return m.cells.String()
}

func (m *model) updateBoids() {
	m.previousBoids = append(m.previousBoids[:0], m.boids...)

	if len(m.previousBoids) < spatialGridBoidThreshold {
		for i := range m.boids {
			m.boids[i].update(m.previousBoids)
		}
	} else {
		m.nearbyGrid.cellSize = radius
		m.nearbyGrid.rebuild(m.previousBoids)
		if cap(m.candidateIndexes) < len(m.previousBoids) {
			m.candidateIndexes = make([]int, 0, len(m.previousBoids))
		}

		for i := range m.boids {
			m.candidateIndexes = m.nearbyGrid.candidateIndexes(m.previousBoids[i].pos, m.candidateIndexes)
			m.boids[i].updateWithCandidateIndexes(m.previousBoids, m.candidateIndexes)
		}
	}

	for i := range m.boids {
		m.boids[i].move()
		drawTriangle(&m.cells, m.boids[i].pos, m.boids[i].forward)
	}
}

func drawTriangle(cb *cellbuffer, centre, dir Point) {
	cb.set(int(centre.x), int(centre.y), triangleRuneTable[dir])
}

var triangleRuneTable = map[Point]rune{
	{-1, -1}: '◤',
	{-1, 0}:  '◀',
	{-1, 1}:  '◣',
	{0, 1}:   '▼',
	{1, 1}:   '◢',
	{1, 0}:   '▶',
	{1, -1}:  '◥',
	{0, -1}:  '▲',
}
