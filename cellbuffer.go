package main

import (
	"strings"
	"unicode/utf8"
)

type cellbuffer struct {
	cells       []rune
	widthCells  int
	heightCells int
	dirty       []int
	dirtyMark   []uint32
	dirtyFrame  uint32
}

func (c *cellbuffer) init(w, h int) {
	if w == 0 {
		return
	}
	c.widthCells = w
	c.heightCells = h
	c.cells = make([]rune, w*h)
	c.dirty = c.dirty[:0]
	c.dirtyFrame++
	if c.dirtyFrame == 0 {
		c.dirtyFrame = 1
	}
	c.dirtyMark = make([]uint32, w*h)
	c.wipe()
}

func (c *cellbuffer) set(x, y int, r rune) {
	if x < 0 || y < 0 || x >= c.widthCells || y >= c.heightCells {
		return
	}
	i := y*c.widthCells + x
	if i >= len(c.cells) {
		return
	}
	if c.dirtyMark[i] != c.dirtyFrame {
		c.dirtyMark[i] = c.dirtyFrame
		c.dirty = append(c.dirty, i)
	}
	c.cells[i] = r
}

func (c *cellbuffer) wipe() {
	if len(c.dirty) == 0 {
		for i := range c.cells {
			c.cells[i] = ' '
		}
		return
	}

	for _, i := range c.dirty {
		c.cells[i] = ' '
	}
	c.dirty = c.dirty[:0]
	c.dirtyFrame++
	if c.dirtyFrame == 0 {
		for i := range c.dirtyMark {
			c.dirtyMark[i] = 0
		}
		c.dirtyFrame = 1
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

	if len(c.cells) == 0 || c.widthCells == 0 {
		return b.String()
	}

	width := c.widthCells
	for rowStart := 0; rowStart < len(c.cells); rowStart += width {
		rowEnd := rowStart + width
		if rowStart > 0 {
			b.WriteByte('\n')
		}

		for i := rowStart; i < rowEnd; {
			if c.cells[i] != ' ' {
				if c.cells[i] < utf8.RuneSelf {
					b.WriteByte(byte(c.cells[i]))
				} else {
					b.WriteRune(c.cells[i])
				}
				i++
				continue
			}

			runEnd := i + 1
			for runEnd < rowEnd && c.cells[runEnd] == ' ' {
				runEnd++
			}
			writeSpaces(&b, runEnd-i)
			i = runEnd
		}
	}
	return b.String()
}

const spaces = "                                                                "

func writeSpaces(b *strings.Builder, count int) {
	for count > len(spaces) {
		b.WriteString(spaces)
		count -= len(spaces)
	}
	b.WriteString(spaces[:count])
}
