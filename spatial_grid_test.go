package main

import (
	"math"
	"slices"
	"testing"
)

func TestSpatialGridCellFor(t *testing.T) {
	grid := newSpatialGrid(10, 100, 100)

	tests := []struct {
		name string
		pos  Point
		want spatialGridCell
	}{
		{
			name: "origin",
			pos:  Point{x: 0, y: 0},
			want: spatialGridCell{x: 0, y: 0},
		},
		{
			name: "positive fractional before boundary",
			pos:  Point{x: 9.9, y: 19.5},
			want: spatialGridCell{x: 0, y: 1},
		},
		{
			name: "positive boundary",
			pos:  Point{x: 10, y: 20},
			want: spatialGridCell{x: 1, y: 2},
		},
		{
			name: "negative fractional",
			pos:  Point{x: -0.1, y: -10.1},
			want: spatialGridCell{x: 0, y: 0},
		},
		{
			name: "negative boundary",
			pos:  Point{x: -10, y: -20},
			want: spatialGridCell{x: 0, y: 0},
		},
		{
			name: "positive out of bounds",
			pos:  Point{x: 105, y: 105},
			want: spatialGridCell{x: 9, y: 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grid.cellFor(tt.pos); got != tt.want {
				t.Fatalf("cellFor(%+v) = %+v, want %+v", tt.pos, got, tt.want)
			}
		})
	}
}

func TestSpatialGridCellForFractionalCellSize(t *testing.T) {
	grid := newSpatialGrid(2.5, 100, 100)

	tests := []struct {
		name string
		pos  Point
		want spatialGridCell
	}{
		{
			name: "positive fractional cell",
			pos:  Point{x: 4.9, y: 5},
			want: spatialGridCell{x: 1, y: 2},
		},
		{
			name: "negative fractional cell",
			pos:  Point{x: -0.1, y: -2.6},
			want: spatialGridCell{x: 0, y: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := grid.cellFor(tt.pos); got != tt.want {
				t.Fatalf("cellFor(%+v) = %+v, want %+v", tt.pos, got, tt.want)
			}
		})
	}
}

func TestSpatialGridInvalidCellSizeFallsBack(t *testing.T) {
	tests := []struct {
		name     string
		cellSize float64
	}{
		{
			name:     "zero",
			cellSize: 0,
		},
		{
			name:     "negative",
			cellSize: -5,
		},
		{
			name:     "positive infinity",
			cellSize: math.Inf(1),
		},
		{
			name:     "negative infinity",
			cellSize: math.Inf(-1),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grid := newSpatialGrid(tt.cellSize, 100, 100)

			if got, want := grid.cellSize, float64(defaultSpatialGridCellSize); got != want {
				t.Fatalf("cellSize = %v, want %v", got, want)
			}
			if got, want := grid.cellFor(Point{x: 1.2, y: -0.1}), (spatialGridCell{x: 1, y: 0}); got != want {
				t.Fatalf("cellFor() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestSpatialGridCandidateIndexes(t *testing.T) {
	grid := newSpatialGrid(10, 100, 100)
	grid.rebuild([]boid{
		{pos: Point{x: 5, y: 5}},     // same cell
		{pos: Point{x: 9.9, y: 9.9}}, // same cell
		{pos: Point{x: 10, y: 5}},    // adjacent
		{pos: Point{x: 19.9, y: 19.9}},
		{pos: Point{x: 20, y: 5}}, // non-adjacent
		{pos: Point{x: 5, y: 20}}, // non-adjacent
	})

	got := grid.candidateIndexes(Point{x: 5, y: 5}, nil)
	assertSameIndexes(t, got, []int{0, 1, 2, 3})
}

func TestSpatialGridCandidateIndexesUsesDestination(t *testing.T) {
	grid := newSpatialGrid(10, 100, 100)
	grid.rebuild([]boid{
		{pos: Point{x: 5, y: 5}},
		{pos: Point{x: 15, y: 5}},
	})
	dst := []int{99, 100, 101}

	got := grid.candidateIndexes(Point{x: 5, y: 5}, dst)

	assertSameIndexes(t, got, []int{0, 1})
}

func TestSpatialGridRebuildClearsStaleBuckets(t *testing.T) {
	grid := newSpatialGrid(10, 100, 100)
	grid.rebuild([]boid{
		{pos: Point{x: 5, y: 5}},
		{pos: Point{x: 15, y: 5}},
	})

	grid.rebuild([]boid{
		{pos: Point{x: 55, y: 55}},
	})

	got := grid.candidateIndexes(Point{x: 5, y: 5}, nil)
	if len(got) != 0 {
		t.Fatalf("candidateIndexes() near stale bucket = %v, want empty", got)
	}

	got = grid.candidateIndexes(Point{x: 55, y: 55}, got)
	assertSameIndexes(t, got, []int{0})
}

func TestSpatialGridRebuildKeepsFixedLayoutAcrossFrames(t *testing.T) {
	grid := newSpatialGrid(10, 100, 50)
	if got, want := grid.cellCols, 10; got != want {
		t.Fatalf("initial cellCols = %d, want %d", got, want)
	}
	if got, want := grid.cellRows, 5; got != want {
		t.Fatalf("initial cellRows = %d, want %d", got, want)
	}

	grid.rebuild(nil)
	if got, want := len(grid.cells), 50; got != want {
		t.Fatalf("len(grid.cells) after empty rebuild = %d, want %d", got, want)
	}

	grid.rebuild([]boid{
		{pos: Point{x: 1, y: 1}},
	})
	if got, want := len(grid.cells), 50; got != want {
		t.Fatalf("len(grid.cells) after non-empty rebuild = %d, want %d", got, want)
	}

	// Candidate lookups should still be bounded by the fixed world grid.
	got := grid.candidateIndexes(Point{x: 99, y: 49}, nil)
	if len(got) != 0 {
		t.Fatalf("candidateIndexes() at far edge = %v, want empty", got)
	}
}

func assertSameIndexes(t *testing.T, got, want []int) {
	t.Helper()

	got = append([]int(nil), got...)
	want = append([]int(nil), want...)
	slices.Sort(got)
	slices.Sort(want)

	if !slices.Equal(got, want) {
		t.Fatalf("indexes = %v, want %v", got, want)
	}
}
