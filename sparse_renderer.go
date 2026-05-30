package main

import (
	"io"
	"strconv"
	"unicode/utf8"
)

type sparseRenderer struct {
	widthCells    int
	heightCells   int
	prevGlyphs    []rune
	nextGlyphs    []rune
	prevOccupied  []bool
	nextFrameMark []uint32
	prevTouched   []int
	nextTouched   []int
	frameMark     uint32
	output        []byte
	previousStat  string
	initialized   bool
}

func (s *sparseRenderer) reset(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}

	s.widthCells = width
	s.heightCells = height

	cellCount := width * height
	s.resizeCellState(cellCount)
	clear(s.prevGlyphs)
	clear(s.nextGlyphs)
	clear(s.prevOccupied)
	clear(s.nextFrameMark)
	s.prevTouched = s.prevTouched[:0]
	s.nextTouched = s.nextTouched[:0]
	s.frameMark = 0

	s.previousStat = ""
	s.initialized = false
}

func (s *sparseRenderer) render(w io.Writer, boids []boid, status string) error {
	width := s.widthCells
	height := s.heightCells
	s.prepareNextFrame(len(boids))

	for i := range boids {
		x := int(boids[i].pos.x)
		y := int(boids[i].pos.y)
		if x < 0 || y < 0 || x >= width || y >= height {
			continue
		}
		s.setNextCell(cellIndex(width, x, y), triangleRune(boids[i].vel))
	}
	return s.flushFrame(w, status)
}

func (s *sparseRenderer) renderSimulation(w io.Writer, sim *simulation, status string) error {
	return s.renderVectors(w, sim.x, sim.y, sim.vx, sim.vy, status)
}

func (s *sparseRenderer) renderVectors(w io.Writer, xs, ys, vxs, vys []float64, status string) error {
	width := s.widthCells
	height := s.heightCells
	s.prepareNextFrame(len(xs))

	for i := range xs {
		x := int(xs[i])
		y := int(ys[i])
		if x < 0 || y < 0 || x >= width || y >= height {
			continue
		}
		s.setNextCell(cellIndex(width, x, y), triangleRune(Point{x: vxs[i], y: vys[i]}))
	}

	return s.flushFrame(w, status)
}

func (s *sparseRenderer) prepareNextFrame(boidCount int) {
	cellCount := s.widthCells * s.heightCells
	if len(s.prevGlyphs) != cellCount {
		s.resizeCellState(cellCount)
		clear(s.prevGlyphs)
		clear(s.nextGlyphs)
		clear(s.prevOccupied)
		clear(s.nextFrameMark)
		s.prevTouched = s.prevTouched[:0]
	}
	if cap(s.nextTouched) < boidCount {
		s.nextTouched = make([]int, 0, boidCount)
	} else {
		s.nextTouched = s.nextTouched[:0]
	}
	s.frameMark++
	if s.frameMark == 0 {
		clear(s.nextFrameMark)
		s.frameMark = 1
	}
}

func (s *sparseRenderer) resizeCellState(cellCount int) {
	if cap(s.prevGlyphs) < cellCount {
		s.prevGlyphs = make([]rune, cellCount)
	} else {
		s.prevGlyphs = s.prevGlyphs[:cellCount]
	}
	if cap(s.nextGlyphs) < cellCount {
		s.nextGlyphs = make([]rune, cellCount)
	} else {
		s.nextGlyphs = s.nextGlyphs[:cellCount]
	}
	if cap(s.prevOccupied) < cellCount {
		s.prevOccupied = make([]bool, cellCount)
	} else {
		s.prevOccupied = s.prevOccupied[:cellCount]
	}
	if cap(s.nextFrameMark) < cellCount {
		s.nextFrameMark = make([]uint32, cellCount)
	} else {
		s.nextFrameMark = s.nextFrameMark[:cellCount]
	}
}

func (s *sparseRenderer) setNextCell(index int, glyph rune) {
	if s.nextFrameMark[index] != s.frameMark {
		s.nextFrameMark[index] = s.frameMark
		s.nextTouched = append(s.nextTouched, index)
	}
	s.nextGlyphs[index] = glyph
}

func (s *sparseRenderer) flushFrame(w io.Writer, status string) error {
	width := s.widthCells
	out := s.output[:0]
	for _, i := range s.prevTouched {
		nextOccupied := s.nextFrameMark[i] == s.frameMark
		next := s.nextGlyphs[i]
		if nextOccupied && s.prevGlyphs[i] == next {
			continue
		}
		x, y := cellXY(width, i)
		out = appendCursor(out, x, y)
		if !nextOccupied {
			next = ' '
		}
		out = appendRune(out, next)
	}

	for _, i := range s.nextTouched {
		if s.prevOccupied[i] {
			continue
		}
		x, y := cellXY(width, i)
		out = appendCursor(out, x, y)
		out = appendRune(out, s.nextGlyphs[i])
	}

	for _, i := range s.prevTouched {
		if s.nextFrameMark[i] != s.frameMark {
			s.prevOccupied[i] = false
			s.prevGlyphs[i] = 0
		}
	}
	for _, i := range s.nextTouched {
		s.prevOccupied[i] = true
		s.prevGlyphs[i] = s.nextGlyphs[i]
	}
	s.prevTouched, s.nextTouched = s.nextTouched, s.prevTouched[:0]

	out = s.appendStatus(out, status)
	s.initialized = true
	s.output = out

	if len(out) == 0 {
		return nil
	}
	_, err := w.Write(out)
	return err
}

func (s *sparseRenderer) appendStatus(out []byte, status string) []byte {
	if s.widthCells <= 0 {
		return out
	}

	statusLine := fitStatusLine(status, s.widthCells)
	if !s.initialized || s.previousStat != statusLine {
		out = appendCursor(out, 0, s.heightCells)
		out = append(out, statusLine...)
		out = appendSpacesToBytes(out, s.widthCells-len(statusLine))
		s.previousStat = statusLine
	}

	return out
}

func cellIndex(width, x, y int) int {
	return y*width + x
}

func cellXY(width, index int) (int, int) {
	return index % width, index / width
}

func appendCursor(out []byte, x, y int) []byte {
	// x/y are zero-based coordinates.
	out = append(out, "\x1b["...)
	out = strconv.AppendInt(out, int64(y+1), 10)
	out = append(out, ';')
	out = strconv.AppendInt(out, int64(x+1), 10)
	out = append(out, 'H')
	return out
}

func appendRune(out []byte, r rune) []byte {
	if r < utf8.RuneSelf {
		return append(out, byte(r))
	}

	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return append(out, buf[:n]...)
}

func appendSpacesToBytes(out []byte, count int) []byte {
	for count > len(spaces) {
		out = append(out, spaces...)
		count -= len(spaces)
	}
	return append(out, spaces[:count]...)
}
