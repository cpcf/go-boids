package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHeadlessValidatesOptions(t *testing.T) {
	cfg := defaultConfig()

	tests := []struct {
		name string
		opts headlessOptions
	}{
		{
			name: "negative frames",
			opts: headlessOptions{
				frames: -1,
				width:  80,
				height: 24,
			},
		},
		{
			name: "zero width",
			opts: headlessOptions{
				frames: 1,
				width:  0,
				height: 24,
			},
		},
		{
			name: "negative height",
			opts: headlessOptions{
				frames: 1,
				width:  80,
				height: -24,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := runHeadless(cfg, tt.opts); err == nil {
				t.Fatalf("runHeadless(%+v) = nil error, want error", tt.opts)
			}
		})
	}
}

func TestRunHeadlessZeroFramesDoesNotStep(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(1234)

	opts := headlessOptions{
		frames: 0,
		width:  40,
		height: 25,
	}

	summary, err := runHeadless(cfg, opts)
	if err != nil {
		t.Fatalf("runHeadless() = %v", err)
	}

	sim := newSimulation(cfg)
	sim.Resize(opts.width, opts.height)
	expected := sim.Stats()
	expected.frames = opts.frames

	if summary != expected {
		t.Fatalf("summary = %#v, want %#v", summary, expected)
	}
}

func TestRunHeadlessSeededRunIsStable(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(1)

	opts := headlessOptions{
		frames: 10,
		width:  80,
		height: 24,
	}

	first, err := runHeadless(cfg, opts)
	if err != nil {
		t.Fatalf("first runHeadless() = %v", err)
	}
	second, err := runHeadless(cfg, opts)
	if err != nil {
		t.Fatalf("second runHeadless() = %v", err)
	}

	if first != second {
		t.Fatalf("runHeadless() summaries differ: first=%#v second=%#v", first, second)
	}

	var firstBuffer, secondBuffer bytes.Buffer
	if err := writeHeadlessSummary(&firstBuffer, first); err != nil {
		t.Fatalf("writeHeadlessSummary(first) = %v", err)
	}
	if err := writeHeadlessSummary(&secondBuffer, second); err != nil {
		t.Fatalf("writeHeadlessSummary(second) = %v", err)
	}

	if firstBuffer.String() != secondBuffer.String() {
		t.Fatalf("format output differs: first=%q second=%q", firstBuffer.String(), secondBuffer.String())
	}

	want := "frames=10 width=80 height=24 boids=26 avg_speed=0.919 min_speed=0.210 max_speed=1.290 centroid_x=40.748 centroid_y=10.579\n"
	if got := firstBuffer.String(); got != want {
		t.Fatalf("formatted output = %q, want %q", got, want)
	}
}

func TestWriteHeadlessSummaryFormat(t *testing.T) {
	summary := simulationStats{
		frames:    10,
		width:     80,
		height:    24,
		boids:     26,
		avgSpeed:  0.842123,
		minSpeed:  0.117333,
		maxSpeed:  1.0,
		centroidX: 40.1234,
		centroidY: 12.4567,
	}

	var buf bytes.Buffer
	if err := writeHeadlessSummary(&buf, summary); err != nil {
		t.Fatalf("writeHeadlessSummary() = %v", err)
	}

	got := buf.String()
	want := "frames=10 width=80 height=24 boids=26 avg_speed=0.842 min_speed=0.117 max_speed=1.000 centroid_x=40.123 centroid_y=12.457\n"
	if got != want {
		t.Fatalf("write output = %q, want %q", got, want)
	}

	if !strings.HasSuffix(got, "\n") {
		t.Fatalf("summary should end with newline, got %q", got)
	}
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("summary should contain one newline, got %q", got)
	}
}
