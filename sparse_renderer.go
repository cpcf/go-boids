package main

import (
	"io"
	"strconv"
	"unicode/utf8"
)

type sparseRenderer struct {
	widthCells   int
	heightCells  int
	prevBoids    map[int]rune
	nextBoids    map[int]rune
	output       []byte
	previousStat string
	initialized  bool
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

	s.prevBoids = clearRuneMap(s.prevBoids)
	s.nextBoids = clearRuneMap(s.nextBoids)
	if s.prevBoids == nil {
		s.prevBoids = make(map[int]rune)
	}
	if s.nextBoids == nil {
		s.nextBoids = make(map[int]rune)
	}

	s.previousStat = ""
	s.initialized = false
}

func (s *sparseRenderer) render(w io.Writer, boids []boid, status string) error {
	width := s.widthCells
	height := s.heightCells
	if s.prevBoids == nil {
		s.prevBoids = make(map[int]rune, len(boids))
	}
	if s.nextBoids == nil {
		s.nextBoids = make(map[int]rune, len(boids))
	}

	s.nextBoids = clearRuneMap(s.nextBoids)
	for _, boid := range boids {
		x := int(boid.pos.x)
		y := int(boid.pos.y)
		if x < 0 || y < 0 || x >= width || y >= height {
			continue
		}
		s.nextBoids[cellIndex(width, x, y)] = triangleRune(boid.vel)
	}

	out := s.output[:0]
	for i, prev := range s.prevBoids {
		next, ok := s.nextBoids[i]
		if ok && prev == next {
			continue
		}
		x, y := cellXY(width, i)
		out = appendCursor(out, x, y)
		if !ok {
			next = ' '
		}
		out = appendRune(out, next)
	}

	for i, next := range s.nextBoids {
		if _, ok := s.prevBoids[i]; ok {
			continue
		}
		x, y := cellXY(width, i)
		out = appendCursor(out, x, y)
		out = appendRune(out, next)
	}

	s.prevBoids, s.nextBoids = s.nextBoids, s.prevBoids

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

func clearRuneMap(m map[int]rune) map[int]rune {
	for k := range m {
		delete(m, k)
	}
	return m
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
