package main

import (
	"reflect"
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
	m.sim.boids = []boid{{
		pos:           Point{x: 1, y: 1},
		vel:           Point{x: 1, y: 0},
		maxX:          20,
		maxY:          10,
		bounce:        false,
		clampMinSpeed: false,
		forward:       Point{x: 1, y: 0},
	}}

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
		cfg: cfg,
		sim: newSimulation(cfg),
	}
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)
	beforeBoids := copyBoidsForTest(m.sim.Boids())

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
