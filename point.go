package main

import "math"

type Point struct {
	x float64
	y float64
}

// Adds the coordinates of two points
func (p1 Point) Add(p2 Point) Point {
	return Point{x: p1.x + p2.x, y: p1.y + p2.y}
}

// Subtracts the coordinates of two points
func (p1 Point) Subtract(p2 Point) Point {
	return Point{x: p1.x - p2.x, y: p1.y - p2.y}
}

// Multiplies the coordinates of two points
func (p1 Point) Multiply(p2 Point) Point {
	return Point{x: p1.x * p2.x, y: p1.y * p2.y}
}

// Adds a scalar value to the coordinates of a point
func (point Point) AddV(val float64) Point {
	return Point{x: point.x + val, y: point.y + val}
}

// Multiplies the coordinates of a point by a scalar value
func (point Point) MultiplyV(val float64) Point {
	return Point{x: point.x * val, y: point.y * val}
}

// Divides the coordinates of a point by a scalar value
func (point Point) DivideV(val float64) Point {
	return Point{x: point.x / val, y: point.y / val}
}

// Clamps the coordinates of a point to the specified lower and upper bounds
func (point Point) Limit(lower, upper float64) Point {
	return Point{x: math.Min(math.Max(point.x, lower), upper), y: math.Min(math.Max(point.y, lower), upper)}
}

// Calculates the distance between two points
func (p1 Point) Distance(p2 Point) float64 {
	return math.Sqrt(p1.distanceSquared(p2))
}

func (p1 Point) distanceSquared(p2 Point) float64 {
	return p1.Subtract(p2).magnitudeSquared()
}

func (point Point) magnitudeSquared() float64 {
	return point.x*point.x + point.y*point.y
}

// Normalize returns a new point with the same direction but with a magnitude of 1
func (p1 Point) Normalize() Point {
	magnitude := math.Sqrt(p1.magnitudeSquared())
	return Point{x: math.Floor(p1.x/magnitude + 0.5), y: math.Floor(p1.y/magnitude + 0.5)}
}

// Linearly interpolates between two points by a given proportion
func (p1 Point) Lerp(p2 Point, proportion float64) Point {
	return Point{
		x: Lerp(p1.x, p2.x, proportion),
		y: Lerp(p1.y, p2.y, proportion),
	}
}

// Linearly interpolates between two values by a given proportion
func Lerp(p0, p1, t float64) float64 {
	return (1-t)*p0 + t*p1
}
