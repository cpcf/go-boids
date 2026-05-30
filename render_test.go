package main

import (
	"reflect"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestModelSpaceTogglesPaused(t *testing.T) {
	m := model{
		cfg: defaultConfig(),
		sim: newSimulation(defaultConfig()),
	}
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)

	m = modelWithUpdate(t, m, tea.KeyMsg{
		Type:  tea.KeySpace,
		Runes: []rune{' '},
	})
	if !m.paused {
		t.Fatal("space should set paused=true")
	}

	m = modelWithUpdate(t, m, tea.KeyMsg{
		Type:  tea.KeySpace,
		Runes: []rune{' '},
	})
	if m.paused {
		t.Fatal("space should set paused=false")
	}
}

func TestModelHandleInputRuneReturnsQuitWithoutBubbleTeaMessage(t *testing.T) {
	m := model{
		cfg: defaultConfig(),
		sim: newSimulation(defaultConfig()),
	}

	if !m.handleInputRune('q') {
		t.Fatal("q should request quit")
	}

	if m.handleInputRune('s') {
		t.Fatal("s should not request quit")
	}
	if !m.showStats {
		t.Fatal("s should toggle stats")
	}
}

func TestModelHandleWindowResizeInitializesState(t *testing.T) {
	m := model{
		cfg: defaultConfig(),
		sim: newSimulation(defaultConfig()),
	}

	m.handleWindowResize(20, 10)

	if got, want := m.cells.width(), 20; got != want {
		t.Fatalf("cells width = %d, want %d", got, want)
	}
	if got, want := m.cells.height(), 9; got != want {
		t.Fatalf("cells height = %d, want %d", got, want)
	}
	if got, want := m.sim.width, 20; got != want {
		t.Fatalf("simulation width = %d, want %d", got, want)
	}
	if got, want := m.sim.height, 9; got != want {
		t.Fatalf("simulation height = %d, want %d", got, want)
	}
	if !m.cells.ready() {
		t.Fatal("cells should be ready after resize")
	}
}

func TestModelFrameDoesNotStepWhenPaused(t *testing.T) {
	m := model{
		cfg:    defaultConfig(),
		paused: true,
	}
	m.sim = newSimulation(defaultConfig())
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)
	m.drawBoids()

	before := copyBoidsForTest(m.sim.Boids())

	m = modelWithUpdate(t, m, frameMsg{})

	after := m.sim.Boids()
	if !boidsEqualForTest(before, after) {
		t.Fatal("frame update while paused should not step simulation")
	}
}

func TestModelSingleStepOnlyWhenPaused(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(123)
	m := model{
		cfg:    cfg,
		paused: true,
	}
	m.cells.init(20, 10)
	m.sim = simulation{cfg: cfg}
	m.sim.Resize(20, 10)
	m.sim.setBoids([]boid{{
		pos:           Point{x: 1, y: 1},
		vel:           Point{x: 1, y: 0},
		maxX:          20,
		maxY:          10,
		bounce:        false,
		clampMinSpeed: false,
	}})

	beforePaused := copyBoidsForTest(m.sim.Boids())
	m = modelWithUpdate(t, m, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'.'},
	})
	if boidsEqualForTest(beforePaused, m.sim.Boids()) {
		t.Fatal("single-step should advance simulation when paused")
	}

	m.paused = false
	beforeRunning := copyBoidsForTest(m.sim.Boids())
	m = modelWithUpdate(t, m, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'.'},
	})
	if !boidsEqualForTest(beforeRunning, m.sim.Boids()) {
		t.Fatal("single-step should not advance simulation when not paused")
	}
}

func TestTriangleRuneFromVelocity(t *testing.T) {
	tests := []struct {
		name string
		dir  Point
		want rune
	}{
		{
			name: "diagonal positive",
			dir:  Point{x: 3, y: 4},
			want: '◢',
		},
		{
			name: "diagonal negative",
			dir:  Point{x: -3, y: -4},
			want: '◤',
		},
		{
			name: "horizontal",
			dir:  Point{x: 4, y: 0},
			want: '▶',
		},
		{
			name: "horizontal negative",
			dir:  Point{x: -7, y: 0},
			want: '◀',
		},
		{
			name: "vertical down",
			dir:  Point{x: 0, y: 7},
			want: '▼',
		},
		{
			name: "vertical up",
			dir:  Point{x: 0, y: -2},
			want: '▲',
		},
		{
			name: "diagonal southwest",
			dir:  Point{x: -4, y: -5},
			want: '◤',
		},
		{
			name: "small component rounds to zero",
			dir:  Point{x: 5, y: 1},
			want: '▶',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := triangleRune(tt.dir); got != tt.want {
				t.Fatalf("triangleRune(%+v) = %c, want %c", tt.dir, got, tt.want)
			}
		})
	}
}

func TestModelResetReinitializesBoidsForCurrentSize(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(1234)
	m := model{cfg: cfg}
	m.cells.init(30, 20)
	m.sim = newSimulation(cfg)
	m.sim.Resize(12, 8)
	m.resetSimulation()
	initialView := m.View()

	initial := copyBoidsForTest(m.sim.Boids())

	m.sim.Step()
	m.drawBoids()
	steppedView := m.View()
	if boidsEqualForTest(initial, m.sim.Boids()) {
		t.Fatal("expected simulation state to change after step")
	}
	if steppedView == initialView {
		t.Fatal("expected view to change after stepping")
	}

	m = modelWithUpdate(t, m, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'r'},
	})

	after := m.sim.Boids()
	if !boidsEqualForTest(initial, after) {
		t.Fatalf("resetSimulation should restore initial boids for current size/config; got=%#v want=%#v", after, initial)
	}
	if m.View() != initialView {
		t.Fatal("reset should redraw immediately with reset boids")
	}

	if m.sim.width != m.cells.width() || m.sim.height != m.cells.height() {
		t.Fatalf("simulation size = (%d, %d), want (%d, %d)", m.sim.width, m.sim.height, m.cells.width(), m.cells.height())
	}

	m.resetSimulation()
	if !m.cells.ready() {
		t.Fatal("cells should remain ready after reset")
	}
}

func TestModelQuitKeysQuit(t *testing.T) {
	tests := []tea.KeyMsg{
		{
			Type:  tea.KeyRunes,
			Runes: []rune{'q'},
		},
		{
			Type: tea.KeyEsc,
		},
		{
			Type:  tea.KeyCtrlC,
			Runes: []rune{},
		},
	}

	for _, key := range tests {
		m := model{}
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("Update(%v) = nil cmd, want tea.Quit", key)
		}

		msg := cmd()
		if !isQuitMsgForTest(msg) {
			t.Fatalf("Update(%v) cmd() = %T, want tea.QuitMsg", key, msg)
		}
	}
}

func TestModelOtherKeysAreIgnored(t *testing.T) {
	cfg := defaultConfig()
	m := model{
		cfg:       cfg,
		sim:       newSimulation(cfg),
		showStats: true,
	}
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)
	beforeBoids := copyBoidsForTest(m.sim.Boids())
	beforeShowStats := m.showStats

	updated, cmd := m.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'x'},
	})
	if cmd != nil {
		t.Fatalf("Update(unhandled key) returned cmd %T, want nil", cmd)
	}
	got := updated.(model)
	if got.paused != m.paused {
		t.Fatalf("paused = %v, want %v", got.paused, m.paused)
	}
	if got.showStats != beforeShowStats {
		t.Fatalf("showStats = %v, want %v", got.showStats, beforeShowStats)
	}
	if got.cfg != m.cfg {
		t.Fatalf("cfg changed for unknown key: got %+v want %+v", got.cfg, m.cfg)
	}
	if got.sim.cfg != m.sim.cfg {
		t.Fatalf("sim cfg changed for unknown key: got %+v want %+v", got.sim.cfg, m.sim.cfg)
	}
	if !boidsEqualForTest(beforeBoids, got.sim.Boids()) {
		t.Fatal("boids changed for unknown key")
	}
}

func TestModelStatsKeyTogglesStatusLine(t *testing.T) {
	cfg := defaultConfig()
	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)
	m.drawBoids()

	initialStatus := m.statusLine()
	if !strings.Contains(initialStatus, "s stats") {
		t.Fatalf("initial status line should expose stats toggle: %q", initialStatus)
	}
	if strings.Contains(initialStatus, "frame ") {
		t.Fatalf("initial status line should be help mode: %q", initialStatus)
	}

	updated := modelWithUpdate(t, m, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'s'},
	})
	if !updated.showStats {
		t.Fatal("pressing s should enable stats mode")
	}

	statsStatus := updated.statusLine()
	if !strings.Contains(statsStatus, "| s help | frame ") {
		t.Fatalf("stats status line should include frame field: %q", statsStatus)
	}

	updated = modelWithUpdate(t, updated, tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'s'},
	})
	if updated.showStats {
		t.Fatal("pressing s again should disable stats mode")
	}
	if updated.statusLine() != initialStatus {
		t.Fatalf("toggled status line should return to help mode: %q vs %q", updated.statusLine(), initialStatus)
	}
}

func TestStatsStatusLineUsesSimulationStats(t *testing.T) {
	stats := simulationStats{
		frames:    42,
		width:     80,
		height:    23,
		boids:     50,
		avgSpeed:  0.833,
		minSpeed:  0.414,
		maxSpeed:  1.2,
		centroidX: 39.24,
		centroidY: 11.44,
	}

	got := statsStatusLine(stats)
	want := "q quit | s help | frame 42 | size 80x23 | boids 50 | speed avg/min/max 0.83/0.41/1.20 | center 39.2,11.4"
	if got != want {
		t.Fatalf("statsStatusLine(%+v) = %q, want %q", stats, got, want)
	}
}

func TestModelRuntimeAdjustmentKeysUpdateModelAndSimulationConfig(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 1.0
	cfg.maxSpeed = 0.2

	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)
	beforeBoids := copyBoidsForTest(m.sim.Boids())

	tests := []struct {
		name         string
		key          rune
		wantRadius   float64
		wantMaxSpeed float64
	}{
		{
			name:         "decrease radius",
			key:          '[',
			wantRadius:   0.5,
			wantMaxSpeed: 0.2,
		},
		{
			name:         "decrease radius clamp",
			key:          '[',
			wantRadius:   minRuntimeRadius,
			wantMaxSpeed: 0.2,
		},
		{
			name:         "decrease max speed",
			key:          '-',
			wantRadius:   minRuntimeRadius,
			wantMaxSpeed: 0.1,
		},
		{
			name:         "decrease max speed clamp",
			key:          '-',
			wantRadius:   minRuntimeRadius,
			wantMaxSpeed: minRuntimeMaxSpeed,
		},
		{
			name:         "increase radius",
			key:          ']',
			wantRadius:   1.0,
			wantMaxSpeed: minRuntimeMaxSpeed,
		},
		{
			name:         "increase max speed",
			key:          '=',
			wantRadius:   1.0,
			wantMaxSpeed: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m = modelWithUpdate(t, m, tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune{tt.key},
			})
			if !almostEqualFloatForTest(m.cfg.radius, tt.wantRadius) {
				t.Fatalf("model radius = %.6f, want %.6f", m.cfg.radius, tt.wantRadius)
			}
			if !almostEqualFloatForTest(m.cfg.maxSpeed, tt.wantMaxSpeed) {
				t.Fatalf("model maxSpeed = %.6f, want %.6f", m.cfg.maxSpeed, tt.wantMaxSpeed)
			}
			if !almostEqualFloatForTest(m.sim.cfg.radius, tt.wantRadius) {
				t.Fatalf("simulation radius = %.6f, want %.6f", m.sim.cfg.radius, tt.wantRadius)
			}
			if !almostEqualFloatForTest(m.sim.cfg.maxSpeed, tt.wantMaxSpeed) {
				t.Fatalf("simulation maxSpeed = %.6f, want %.6f", m.sim.cfg.maxSpeed, tt.wantMaxSpeed)
			}
			if !almostEqualFloatForTest(m.sim.nearbyGrid.cellSize, tt.wantRadius) {
				t.Fatalf("nearby grid cell size = %.6f, want %.6f", m.sim.nearbyGrid.cellSize, tt.wantRadius)
			}
		})
	}

	if !boidsEqualForTest(beforeBoids, m.sim.Boids()) {
		t.Fatal("runtime adjustments should redraw without stepping boids")
	}
}

func TestModelViewIncludesStatusLine(t *testing.T) {
	m := model{
		cfg: defaultConfig(),
		sim: newSimulation(defaultConfig()),
	}
	m.cells.init(20, 9)
	m.sim.Resize(20, 9)
	m.drawBoids()

	got := m.View()
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("View() returned %q, want multiple lines", got)
	}

	gotStatus := lines[len(lines)-1]
	wantStatus := fitStatusLine(m.statusLine(), 20)
	if gotStatus != wantStatus {
		t.Fatalf("View() status line = %q, want %q", gotStatus, wantStatus)
	}
}

func TestModelWindowSizeReservesStatusLine(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(123)
	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}

	m = modelWithUpdate(t, m, tea.WindowSizeMsg{Width: 20, Height: 10})

	if got, want := m.cells.height(), 9; got != want {
		t.Fatalf("cellbuffer height = %d, want %d", got, want)
	}
	if got, want := m.sim.height, 9; got != want {
		t.Fatalf("simulation height = %d, want %d", got, want)
	}
}

func TestModelWindowSizeOneRowShowsOnlyStatus(t *testing.T) {
	cfg := defaultConfig()
	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}

	m = modelWithUpdate(t, m, tea.WindowSizeMsg{Width: 20, Height: 1})

	if got, want := m.cells.height(), 0; got != want {
		t.Fatalf("cellbuffer height = %d, want %d", got, want)
	}
	if got, want := m.sim.height, 0; got != want {
		t.Fatalf("simulation height = %d, want %d", got, want)
	}
	if got := strings.Count(m.View(), "\n"); got != 0 {
		t.Fatalf("one-row View() should not contain newlines, got %q", m.View())
	}
}

func TestModelViewStatusLineReflectsPausedState(t *testing.T) {
	cfg := defaultConfig()
	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}
	m.cells.init(80, 10)
	m.sim.Resize(80, 10)
	m.drawBoids()

	running := m.View()
	runningStatus := strings.Split(running, "\n")[len(strings.Split(running, "\n"))-1]
	if !strings.Contains(runningStatus, "space pause") {
		t.Fatalf("running status line missing pause hint: %q", runningStatus)
	}
	if strings.Contains(runningStatus, ". step") {
		t.Fatalf("running status line should not include single-step hint: %q", runningStatus)
	}

	m.paused = true
	paused := m.View()
	pausedStatus := strings.Split(paused, "\n")[len(strings.Split(paused, "\n"))-1]
	if !strings.Contains(pausedStatus, "space resume") {
		t.Fatalf("paused status line missing resume hint: %q", pausedStatus)
	}
	if !strings.Contains(pausedStatus, ". step") {
		t.Fatalf("paused status line missing single-step hint: %q", pausedStatus)
	}

	if runningStatus == pausedStatus {
		t.Fatalf("expected status lines to differ between running and paused")
	}
}

func TestStatusLineFitsWidth(t *testing.T) {
	status := "q quit | space pause | s stats | r reset | [/] radius 7.0 | -/= max 1.0"

	full := fitStatusLine(status, len(status))
	if full != status {
		t.Fatalf("expected full width status to stay unchanged, got %q", full)
	}

	truncated := fitStatusLine(status, 12)
	if truncated != "q quit | spa" {
		t.Fatalf("expected truncated status, got %q", truncated)
	}

	empty := fitStatusLine(status, 0)
	if empty != "" {
		t.Fatalf("expected empty status when width <= 0, got %q", empty)
	}
}

func modelWithUpdate(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	updated, _ := m.Update(msg)
	updatedModel, ok := updated.(model)
	if !ok {
		t.Fatalf("Update(%T) returned %T, want model", msg, updated)
	}
	return updatedModel
}

func isQuitMsgForTest(msg tea.Msg) bool {
	_, ok := msg.(tea.QuitMsg)
	return ok
}

func boidsEqualForTest(a, b []boid) bool {
	return reflect.DeepEqual(a, b)
}

func copyBoidsForTest(boids []boid) []boid {
	copied := make([]boid, len(boids))
	copy(copied, boids)
	return copied
}
