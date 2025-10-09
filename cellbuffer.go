package main

import (
	"strings"
	"unicode/utf8"
)

type cellbuffer struct {
	cells       []rune
	widthCells  int
	heightCells int
}

func (c *cellbuffer) init(w, h int) {
	if w == 0 {
		return
	}
	c.widthCells = w
	c.heightCells = h
	c.cells = make([]rune, w*h)
	c.wipe()
}

func (c cellbuffer) set(x, y int, r rune) {
	if x < 0 || y < 0 || x >= c.widthCells || y >= c.heightCells {
		return
	}
	i := y*c.widthCells + x
	if i >= len(c.cells) {
		return
	}
	c.cells[i] = r
}

func (c *cellbuffer) wipe() {
	for i := range c.cells {
		c.cells[i] = ' '
	}
}

func (c cellbuffer) width() int {
	return c.widthCells
}

func (c cellbuffer) height() int {
	return c.heightCells
}

func (c cellbuffer) ready() bool {
	return len(c.cells) > 0
}

func (c cellbuffer) String() string {
	var b strings.Builder
	if len(c.cells) > 0 {
		b.Grow(len(c.cells) + c.heightCells - 1)
	}
	for i := 0; i < len(c.cells); i++ {
		if i > 0 && i%c.widthCells == 0 && i < len(c.cells)-1 {
			b.WriteRune('\n')
		}
		if c.cells[i] < utf8.RuneSelf {
			b.WriteByte(byte(c.cells[i]))
		} else {
			b.WriteRune(c.cells[i])
		}
	}
	return b.String()
}
