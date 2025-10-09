package main

import (
	"math"
	"testing"
)

func TestPointAddSubtract(t *testing.T) {
	p := Point{x: 3.5, y: -2}
	other := Point{x: -1.25, y: 4}

	if got, want := p.Add(other), (Point{x: 2.25, y: 2}); got != want {
		t.Fatalf("Add() = %+v, want %+v", got, want)
	}

	if got, want := p.Subtract(other), (Point{x: 4.75, y: -6}); got != want {
		t.Fatalf("Subtract() = %+v, want %+v", got, want)
	}
}

func TestPointLimit(t *testing.T) {
	tests := []struct {
		name         string
		point        Point
		lower, upper float64
		want         Point
	}{
		{
			name:  "within bounds",
			point: Point{x: 0.25, y: -0.5},
			lower: -1,
			upper: 1,
			want:  Point{x: 0.25, y: -0.5},
		},
		{
			name:  "clamps both axes",
			point: Point{x: 2, y: -3},
			lower: -1,
			upper: 1,
			want:  Point{x: 1, y: -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.point.Limit(tt.lower, tt.upper); got != tt.want {
				t.Fatalf("Limit() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestPointDistance(t *testing.T) {
	got := Point{x: -1, y: 2}.Distance(Point{x: 2, y: 6})
	if !near(got, 5) {
		t.Fatalf("Distance() = %v, want 5", got)
	}
}

func TestPointNormalize(t *testing.T) {
	tests := []struct {
		name  string
		point Point
		want  Point
	}{
		{
			name:  "diagonal positive",
			point: Point{x: 3, y: 4},
			want:  Point{x: 1, y: 1},
		},
		{
			name:  "axis aligned negative",
			point: Point{x: 0, y: -2},
			want:  Point{x: 0, y: -1},
		},
		{
			name:  "small component rounds to zero",
			point: Point{x: 5, y: 1},
			want:  Point{x: 1, y: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.point.Normalize(); got != tt.want {
				t.Fatalf("Normalize() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func near(got, want float64) bool {
	return math.Abs(got-want) < 1e-9
}

func pointNear(got, want Point) bool {
	return near(got.x, want.x) && near(got.y, want.y)
}
