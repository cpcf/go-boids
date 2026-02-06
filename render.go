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
	cfg    config
	sim    simulation
	cells  cellbuffer
	paused bool
}

func (m model) Init() tea.Cmd {
	return animate(m.cfg)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyRunes:
			if len(msg.Runes) == 1 {
				switch msg.Runes[0] {
				case 'q':
					return m, tea.Quit
				case ' ':
					m.paused = !m.paused
					return m, nil
				case '.':
					if m.paused {
						m.stepAndDraw(true)
					}
					return m, nil
				case 'r':
					m.resetSimulation()
					return m, nil
				}
			}
		case tea.KeySpace:
			m.paused = !m.paused
			return m, nil
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.cfg = loadConfig()
		m.sim = newSimulation(m.cfg)
		m.sim.Resize(msg.Width, msg.Height)
		m.cells.init(msg.Width, msg.Height)
		return m, nil
	case frameMsg:
		if !m.cells.ready() {
			return m, nil
		}

		m.stepAndDraw(!m.paused)
		return m, animate(m.cfg)
	default:
		return m, nil
	}
}

func (m model) View() string {
	return m.cells.String()
}

func (m *model) stepAndDraw(step bool) {
	if step {
		m.sim.Step()
	}
	m.drawBoids()
}

func (m *model) drawBoids() {
	m.cells.wipe()
	for _, boid := range m.sim.Boids() {
		drawTriangle(&m.cells, boid.pos, boid.forward)
	}
}

func (m *model) resetSimulation() {
	m.sim = newSimulation(m.cfg)
	m.sim.Resize(m.cells.width(), m.cells.height())
	m.drawBoids()
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
