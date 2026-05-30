package main

import "math"

const defaultSpatialGridCellSize = 1

const noCellLink = -1

type spatialGridCell struct {
	x, y int
}

type spatialGrid struct {
	cellSize float64
	width    int
	height   int
	cells    []int
	cellRows int
	cellCols int
	next     []int
	touched  []int
}

func newSpatialGrid(cellSize float64, width, height int) spatialGrid {
	grid := spatialGrid{
		cellSize: validSpatialGridCellSize(cellSize),
		width:    width,
		height:   height,
	}
	grid.resizeCellArrays()
	return grid
}

func (g *spatialGrid) setCellSize(cellSize float64) {
	cellSize = validSpatialGridCellSize(cellSize)
	if g.cellSize == cellSize {
		return
	}
	g.cellSize = cellSize
	g.resizeCellArrays()
}

func (g *spatialGrid) resize(width, height int) {
	g.width = width
	g.height = height
	g.resizeCellArrays()
}

func (g *spatialGrid) rebuild(boids []boid) {
	g.setCellSize(g.cellSize)

	if len(g.cells) == 0 {
		g.touched = g.touched[:0]
		return
	}

	g.clearTouchedCells()
	if len(boids) == 0 {
		return
	}

	if cap(g.next) < len(boids) {
		g.next = make([]int, len(boids))
	}
	g.next = g.next[:len(boids)]

	for i := range boids {
		cell := g.cellFor(boids[i].pos)
		cellIndex := cell.y*g.cellCols + cell.x
		if g.cells[cellIndex] == noCellLink {
			g.touched = append(g.touched, cellIndex)
		}
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
		if y < 0 || y >= g.cellRows {
			continue
		}

		row := y * g.cellCols
		for x := center.x - 1; x <= center.x+1; x++ {
			if x < 0 || x >= g.cellCols {
				continue
			}
			for node := g.cells[row+x]; node != noCellLink; node = g.next[node] {
				visit(node)
			}
		}
	}
}

func (g *spatialGrid) cellFor(pos Point) spatialGridCell {
	if len(g.cells) == 0 || g.cellCols == 0 || g.cellRows == 0 {
		return spatialGridCell{}
	}

	cellSize := validSpatialGridCellSize(g.cellSize)
	cell := spatialGridCell{
		x: int(math.Floor(pos.x / cellSize)),
		y: int(math.Floor(pos.y / cellSize)),
	}

	if cell.x < 0 {
		cell.x = 0
	} else if cell.x >= g.cellCols {
		cell.x = g.cellCols - 1
	}
	if cell.y < 0 {
		cell.y = 0
	} else if cell.y >= g.cellRows {
		cell.y = g.cellRows - 1
	}

	return cell
}

func (g *spatialGrid) resizeCellArrays() {
	if g.width <= 0 || g.height <= 0 {
		g.cellCols = 0
		g.cellRows = 0
		g.cells = g.cells[:0]
		g.touched = g.touched[:0]
		return
	}

	cellCols := int(math.Ceil(float64(g.width) / g.cellSize))
	cellRows := int(math.Ceil(float64(g.height) / g.cellSize))
	if cellCols <= 0 || cellRows <= 0 {
		g.cellCols = 0
		g.cellRows = 0
		g.cells = g.cells[:0]
		g.touched = g.touched[:0]
		return
	}

	cellCount := cellCols * cellRows
	if cap(g.cells) < cellCount {
		g.cells = make([]int, cellCount)
	} else {
		g.cells = g.cells[:cellCount]
	}
	clearSpatialGridLinks(g.cells)
	g.touched = g.touched[:0]

	g.cellCols = cellCols
	g.cellRows = cellRows
}

func (g *spatialGrid) clearTouchedCells() {
	for _, cellIndex := range g.touched {
		if cellIndex >= 0 && cellIndex < len(g.cells) {
			g.cells[cellIndex] = noCellLink
		}
	}
	g.touched = g.touched[:0]
}

func clearSpatialGridLinks(cells []int) {
	for i := range cells {
		cells[i] = noCellLink
	}
}

func validSpatialGridCellSize(cellSize float64) float64 {
	if cellSize <= 0 || math.IsNaN(cellSize) || math.IsInf(cellSize, 0) {
		return defaultSpatialGridCellSize
	}
	return cellSize
}
