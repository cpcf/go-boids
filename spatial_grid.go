package main

import "math"

const defaultSpatialGridCellSize = 1

type spatialGridCell struct {
	x, y int
}

type spatialGrid struct {
	cellSize float64
	cells    map[spatialGridCell][]int
}

func newSpatialGrid(cellSize float64) spatialGrid {
	return spatialGrid{
		cellSize: validSpatialGridCellSize(cellSize),
		cells:    make(map[spatialGridCell][]int),
	}
}

func (g *spatialGrid) rebuild(boids []boid) {
	g.cellSize = validSpatialGridCellSize(g.cellSize)
	if g.cells == nil {
		g.cells = make(map[spatialGridCell][]int)
	}

	for cell := range g.cells {
		g.cells[cell] = g.cells[cell][:0]
	}

	for i := range boids {
		cell := g.cellFor(boids[i].pos)
		g.cells[cell] = append(g.cells[cell], i)
	}

	for cell, indexes := range g.cells {
		if len(indexes) == 0 {
			delete(g.cells, cell)
		}
	}
}

func (g *spatialGrid) candidateIndexes(pos Point, dst []int) []int {
	dst = dst[:0]
	if g.cells == nil {
		return dst
	}

	center := g.cellFor(pos)
	for y := center.y - 1; y <= center.y+1; y++ {
		for x := center.x - 1; x <= center.x+1; x++ {
			dst = append(dst, g.cells[spatialGridCell{x: x, y: y}]...)
		}
	}
	return dst
}

func (g *spatialGrid) cellFor(pos Point) spatialGridCell {
	cellSize := validSpatialGridCellSize(g.cellSize)
	return spatialGridCell{
		x: int(math.Floor(pos.x / cellSize)),
		y: int(math.Floor(pos.y / cellSize)),
	}
}

func validSpatialGridCellSize(cellSize float64) float64 {
	if cellSize <= 0 || math.IsNaN(cellSize) || math.IsInf(cellSize, 0) {
		return defaultSpatialGridCellSize
	}
	return cellSize
}
