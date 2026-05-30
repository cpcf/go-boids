package main

import "testing"

func TestParseTerminalKey(t *testing.T) {
	tests := []struct {
		name     string
		key      rune
		wantQuit bool
		wantKey  rune
	}{
		{
			name:     "ctrl-c",
			key:      0x03,
			wantQuit: true,
		},
		{
			name:     "escape",
			key:      0x1b,
			wantQuit: true,
		},
		{
			name:    "space",
			key:     ' ',
			wantKey: ' ',
		},
		{
			name:    "runtime key",
			key:     'q',
			wantKey: 'q',
		},
		{
			name:    "single step key",
			key:     '.',
			wantKey: '.',
		},
		{
			name:    "radius decrease key",
			key:     '[',
			wantKey: '[',
		},
		{
			name:    "max speed increase key",
			key:     '=',
			wantKey: '=',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTerminalKey(tt.key)
			if got.quit != tt.wantQuit {
				t.Fatalf("parseTerminalKey(%q) quit=%v, want %v", tt.key, got.quit, tt.wantQuit)
			}
			if got.key != tt.wantKey {
				t.Fatalf("parseTerminalKey(%q) key=%q, want %q", tt.key, got.key, tt.wantKey)
			}
		})
	}
}

func TestApplyParsedInputMapsToModelHandler(t *testing.T) {
	m := model{
		cfg: defaultConfig(),
		sim: newSimulation(defaultConfig()),
	}
	m.cells.init(20, 10)
	m.sim.Resize(20, 10)

	if applyParsedInput(&m, parseTerminalKey(' ')) {
		t.Fatal("space should not request quit")
	}
	if !m.paused {
		t.Fatal("space should toggle paused=true")
	}

	if applyParsedInput(&m, parseTerminalKey('q')) != true {
		t.Fatal("q should request quit")
	}
}

func TestEffectiveRenderFPS(t *testing.T) {
	cfg := defaultConfig()
	if got, want := effectiveRenderFPS(cfg), 60; got != want {
		t.Fatalf("effectiveRenderFPS(default) = %d, want %d", got, want)
	}

	cfg.renderFps = 30
	if got, want := effectiveRenderFPS(cfg), 30; got != want {
		t.Fatalf("effectiveRenderFPS(renderFps) = %d, want %d", got, want)
	}

	cfg.renderFps = 0
	cfg.fps = 90
	if got, want := effectiveRenderFPS(cfg), 90; got != want {
		t.Fatalf("effectiveRenderFPS(fps fallback) = %d, want %d", got, want)
	}

	cfg.fps = 0
	if got, want := effectiveRenderFPS(cfg), 60; got != want {
		t.Fatalf("effectiveRenderFPS(default fallback) = %d, want %d", got, want)
	}
}
