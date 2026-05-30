package main

import (
	"fmt"
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
	cfg       config
	sim       simulation
	cells     cellbuffer
	paused    bool
	showStats bool
}

func (m model) Init() tea.Cmd {
	return animate(m.cfg)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, m.handleKey(msg)
	case tea.WindowSizeMsg:
		m.handleWindowResize(msg.Width, msg.Height)
		return m, nil
	case frameMsg:
		m.handleFrame()
		return m, animate(m.cfg)
	default:
		return m, nil
	}
}

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.Type {
	case tea.KeyEsc, tea.KeyCtrlC:
		return tea.Quit
	case tea.KeySpace:
		m.paused = !m.paused
		return nil
	case tea.KeyRunes:
		if len(msg.Runes) != 1 {
			return nil
		}
		return m.handleKeyRune(msg.Runes[0])
	default:
		return nil
	}
}

func (m *model) handleKeyRune(key rune) tea.Cmd {
	if m.handleInputRune(key) {
		return tea.Quit
	}
	return nil
}

func (m *model) handleInputRune(key rune) bool {
	switch key {
	case 'q':
		return true
	case 's':
		m.showStats = !m.showStats
		return false
	case ' ':
		m.paused = !m.paused
		return false
	case '.':
		if m.paused {
			m.stepAndDraw(true)
		}
		return false
	case 'r':
		m.resetSimulation()
		return false
	}

	if adjustment, ok := runtimeAdjustmentForKey(key); ok {
		m.applyRuntimeAdjustment(adjustment)
		return false
	}

	return false
}

func (m *model) applyRuntimeAdjustment(adjustment runtimeAdjustment) {
	m.cfg = applyRuntimeAdjustment(m.cfg, adjustment)
	m.sim.SetConfig(m.cfg)
	if m.cells.ready() {
		m.stepAndDraw(false)
	}
}

func (m *model) handleWindowResize(width, height int) {
	m.cfg = loadConfig()
	m.sim = newSimulation(m.cfg)
	displayHeight := height - 1
	if displayHeight < 0 {
		displayHeight = 0
	}
	m.sim.Resize(width, displayHeight)
	m.cells.init(width, displayHeight)
}

func (m *model) handleFrame() {
	if !m.cells.ready() {
		return
	}

	m.stepAndDraw(!m.paused)
}

func (m model) View() string {
	status := fitStatusLine(m.statusLine(), m.cells.width())
	view := m.cells.String()
	if view == "" {
		return status
	}
	return view + "\n" + status
}

func (m *model) stepAndDraw(step bool) {
	if step {
		m.sim.Step()
	}
	m.drawBoids()
}

func (m *model) drawBoids() {
	m.cells.wipe()
	for i := range m.sim.x {
		drawTriangle(
			&m.cells,
			Point{x: m.sim.x[i], y: m.sim.y[i]},
			Point{x: m.sim.vx[i], y: m.sim.vy[i]},
		)
	}
}

func (m *model) resetSimulation() {
	m.sim = newSimulation(m.cfg)
	m.sim.Resize(m.cells.width(), m.cells.height())
	m.drawBoids()
}

func drawTriangle(cb *cellbuffer, centre, dir Point) {
	cb.set(int(centre.x), int(centre.y), triangleRune(dir))
}

func (m model) statusLine() string {
	if m.showStats {
		return statsStatusLine(m.sim.Stats())
	}
	return m.helpStatusLine()
}

func (m model) helpStatusLine() string {
	resumeOrPause := "pause"
	extras := " | s stats | r reset | [/] radius %.1f | -/= max %.1f"
	if m.paused {
		resumeOrPause = "resume"
		extras = " | . step | s stats | r reset | [/] radius %.1f | -/= max %.1f"
	}
	return "q quit | space " + resumeOrPause + fmt.Sprintf(extras, m.cfg.radius, m.cfg.maxSpeed)
}

func statsStatusLine(stats simulationStats) string {
	return fmt.Sprintf(
		"q quit | s help | frame %d | size %dx%d | boids %d | speed avg/min/max %.2f/%.2f/%.2f | center %.1f,%.1f",
		stats.frames,
		stats.width,
		stats.height,
		stats.boids,
		stats.avgSpeed,
		stats.minSpeed,
		stats.maxSpeed,
		stats.centroidX,
		stats.centroidY,
	)
}

func fitStatusLine(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(s) <= width {
		return s
	}
	return s[:width]
}

func triangleRune(dir Point) rune {
	x := triangleAxis(dir.x, dir.x*dir.x, dir.y*dir.y)
	y := triangleAxis(dir.y, dir.y*dir.y, dir.x*dir.x)

	switch {
	case x < 0 && y < 0:
		return '◤'
	case x < 0 && y == 0:
		return '◀'
	case x < 0 && y > 0:
		return '◣'
	case x == 0 && y > 0:
		return '▼'
	case x > 0 && y > 0:
		return '◢'
	case x > 0 && y == 0:
		return '▶'
	case x > 0 && y < 0:
		return '◥'
	case x == 0 && y < 0:
		return '▲'
	default:
		return 0
	}
}

func triangleAxis(value, square, otherSquare float64) int {
	if value > 0 {
		if 3*square >= otherSquare {
			return 1
		}
		return 0
	}
	if value < 0 && 3*square > otherSquare {
		return -1
	}
	return 0
}
