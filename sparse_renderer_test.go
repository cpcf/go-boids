package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestSparseRendererInitialFrameWritesCellsAndStatus(t *testing.T) {
	var r sparseRenderer
	r.reset(5, 2)

	boids := []boid{
		{
			pos: Point{x: 0, y: 0},
			vel: Point{x: 1, y: 0},
		},
		{
			pos: Point{x: 3, y: 1},
			vel: Point{x: 0, y: 1},
		},
	}

	var buf bytes.Buffer
	if err := r.render(&buf, boids, "stat"); err != nil {
		t.Fatalf("render() = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "\x1b[1;1H▶") {
		t.Fatalf("first frame should render top-left boid, got %q", got)
	}
	if !strings.Contains(got, "\x1b[2;4H▼") {
		t.Fatalf("first frame should render bottom-right boid, got %q", got)
	}
	if !strings.Contains(got, "\x1b[3;1Hstat") {
		t.Fatalf("first frame should render status line, got %q", got)
	}
}

func TestSparseRendererSkipsUnchangedFrames(t *testing.T) {
	var r sparseRenderer
	r.reset(4, 2)
	boids := []boid{{pos: Point{x: 1, y: 0}, vel: Point{x: 1, y: 0}}}

	var buf bytes.Buffer
	if err := r.render(&buf, boids, "stats"); err != nil {
		t.Fatalf("initial render() = %v", err)
	}

	buf.Reset()
	if err := r.render(&buf, boids, "stats"); err != nil {
		t.Fatalf("second render() = %v", err)
	}
	if got := buf.String(); got != "" {
		t.Fatalf("unchanged render should not write output, got %q", got)
	}
}

func TestSparseRendererClearsVacatedCells(t *testing.T) {
	var r sparseRenderer
	r.reset(3, 1)

	var buf bytes.Buffer
	first := []boid{{pos: Point{x: 0, y: 0}, vel: Point{x: 1, y: 0}}}
	if err := r.render(&buf, first, "status"); err != nil {
		t.Fatalf("first render() = %v", err)
	}

	buf.Reset()
	second := []boid{{pos: Point{x: 1, y: 0}, vel: Point{x: 1, y: 0}}}
	if err := r.render(&buf, second, "status"); err != nil {
		t.Fatalf("second render() = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "\x1b[1;1H ") {
		t.Fatalf("moved boid should clear previous cell, got %q", got)
	}
	if !strings.Contains(got, "\x1b[1;2H▶") {
		t.Fatalf("moved boid should draw new cell, got %q", got)
	}
}

func TestSparseRendererUpdatesStatusLineWithoutCellChanges(t *testing.T) {
	var r sparseRenderer
	r.reset(5, 1)
	boids := []boid{{pos: Point{x: 1, y: 0}, vel: Point{x: 1, y: 0}}}

	var buf bytes.Buffer
	if err := r.render(&buf, boids, "abcde"); err != nil {
		t.Fatalf("initial render() = %v", err)
	}

	buf.Reset()
	if err := r.render(&buf, boids, "bb"); err != nil {
		t.Fatalf("second render() = %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "\x1b[2;1Hbb   ") {
		t.Fatalf("status change should update status row, got %q", got)
	}
}

func TestSparseRendererLastBoidWinsCollision(t *testing.T) {
	var r sparseRenderer
	r.reset(2, 1)

	var buf bytes.Buffer
	boids := []boid{
		{pos: Point{x: 0, y: 0}, vel: Point{x: 1, y: 0}},
		{pos: Point{x: 0, y: 0}, vel: Point{x: 0, y: 1}},
	}

	if err := r.render(&buf, boids, "status"); err != nil {
		t.Fatalf("render() = %v", err)
	}

	got := buf.String()
	if strings.Count(got, "▶") != 0 {
		t.Fatalf("earlier boid should be overwritten in final frame, got %q", got)
	}
	if strings.Count(got, "▼") != 1 {
		t.Fatalf("later boid should win collision, got %q", got)
	}
}
