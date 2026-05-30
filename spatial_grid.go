package main

import "math"

const defaultSpatialGridCellSize = 1

const noCellLink = -1

type spatialGridCell struct {
	x, y int
}

type spatialGrid struct {
	cellSize  float64
	cells     []int
	cellRows  int
	cellCols  int
	minCellX  int
	minCellY  int
	next      []int
	boidCellX []int
	boidCellY []int
}

func newSpatialGrid(cellSize float64) spatialGrid {
	return spatialGrid{
		cellSize: validSpatialGridCellSize(cellSize),
	}
}

func (g *spatialGrid) rebuild(boids []boid) {
	g.cellSize = validSpatialGridCellSize(g.cellSize)
	if len(boids) == 0 {
		g.cells = g.cells[:0]
		return
	}

	if cap(g.boidCellX) < len(boids) {
		g.boidCellX = make([]int, len(boids))
		g.boidCellY = make([]int, len(boids))
	}
	g.boidCellX = g.boidCellX[:len(boids)]
	g.boidCellY = g.boidCellY[:len(boids)]

	firstCell := g.cellFor(boids[0].pos)
	minX := firstCell.x
	maxX := firstCell.x
	minY := firstCell.y
	maxY := firstCell.y

	for i := range boids {
		cell := g.cellFor(boids[i].pos)
		g.boidCellX[i] = cell.x
		g.boidCellY[i] = cell.y
		if cell.x < minX {
			minX = cell.x
		}
		if cell.x > maxX {
			maxX = cell.x
		}
		if cell.y < minY {
			minY = cell.y
		}
		if cell.y > maxY {
			maxY = cell.y
		}
	}

	cellCols := maxX - minX + 1
	cellRows := maxY - minY + 1
	if cellCols <= 0 || cellRows <= 0 {
		g.cells = g.cells[:0]
		return
	}

	cellCount := cellCols * cellRows
	if cap(g.cells) < cellCount {
		g.cells = make([]int, cellCount)
	} else {
		g.cells = g.cells[:cellCount]
	}
	for i := range g.cells {
		g.cells[i] = noCellLink
	}

	if cap(g.next) < len(boids) {
		g.next = make([]int, len(boids))
	}
	g.next = g.next[:len(boids)]

	g.minCellX = minX
	g.minCellY = minY
	g.cellCols = cellCols
	g.cellRows = cellRows

	for i := range boids {
		cellX := g.boidCellX[i] - g.minCellX
		cellY := g.boidCellY[i] - g.minCellY
		cellIndex := cellY*g.cellCols + cellX
		g.next[i] = g.cells[cellIndex]
		g.cells[cellIndex] = i
	}
}

func (g *spatialGrid) candidateIndexes(pos Point, dst []int) []int {
	dst = dst[:0]
	if len(g.cells) == 0 {
		return dst
	}

	g.visitNearbyBoidIndexes(pos, func(i int) {
		dst = append(dst, i)
	})
	return dst
}

func (g *spatialGrid) visitNearbyBoidIndexes(pos Point, visit func(int)) {
	if len(g.cells) == 0 {
		return
	}

	center := g.cellFor(pos)
	for y := center.y - 1; y <= center.y+1; y++ {
		if y < g.minCellY || y >= g.minCellY+g.cellRows {
			continue
		}

		row := (y - g.minCellY) * g.cellCols
		for x := center.x - 1; x <= center.x+1; x++ {
			if x < g.minCellX || x >= g.minCellX+g.cellCols {
				continue
			}
			for node := g.cells[row+(x-g.minCellX)]; node != noCellLink; node = g.next[node] {
				visit(node)
			}
		}
	}
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
