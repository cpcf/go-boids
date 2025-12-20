package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type frameMsg struct{}

func animate(cfg config) tea.Cmd {
	if cfg == (config{}) {
		cfg = defaultConfig()
	}

	return tea.Tick(time.Second/time.Duration(cfg.fps), func(_ time.Time) tea.Msg {
		return frameMsg{}
	})
}

type model struct {
	cfg   config
	sim   simulation
	cells cellbuffer
	boids []boid
}

func (m model) Init() tea.Cmd {
	return animate(m.cfg)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.cfg = loadConfig()
		m.sim = newSimulation(m.cfg)
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
	m.sim.Step(m.boids)

	for i := range m.boids {
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
