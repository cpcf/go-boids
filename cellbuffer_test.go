package main

import "testing"

func TestCellbufferSetBoundsAndString(t *testing.T) {
	var c cellbuffer
	c.init(3, 2)

	c.set(0, 0, 'A')
	c.set(2, 1, 'Z')

	c.set(-1, 0, 'x')
	c.set(0, -1, 'x')
	c.set(3, 0, 'x')
	c.set(0, 2, 'x')

	if got, want := c.String(), "A  \n  Z"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestCellbufferWipe(t *testing.T) {
	var c cellbuffer
	c.init(4, 2)
	c.set(0, 0, 'A')
	c.set(3, 1, 'Z')

	c.wipe()

	if got, want := c.String(), "    \n    "; got != want {
		t.Fatalf("String() after wipe = %q, want %q", got, want)
	}
}

func TestCellbufferWipeFallsBackToFullClearForUntouchedCells(t *testing.T) {
	var c cellbuffer
	c.init(4, 1)
	for i := range c.cells {
		c.cells[i] = '.'
	}

	c.wipe()

	if got, want := c.String(), "    "; got != want {
		t.Fatalf("String() after fallback wipe = %q, want %q", got, want)
	}
}

func TestCellbufferWipeClearsOnlyTrackedCells(t *testing.T) {
	var c cellbuffer
	c.init(5, 1)
	c.set(1, 0, 'A')
	c.cells[0] = 'X'
	c.cells[4] = 'B'

	c.wipe()

	if got, want := c.String(), "X   B"; got != want {
		t.Fatalf("String() after partial wipe = %q, want %q", got, want)
	}
}

func TestCellbufferInitZeroWidthLeavesBufferUnready(t *testing.T) {
	var c cellbuffer
	c.init(0, 2)

	if c.ready() {
		t.Fatal("ready() = true, want false")
	}
	if got := c.String(); got != "" {
		t.Fatalf("String() = %q, want empty string", got)
	}
}

func TestCellbufferStringPreservesTriangleGlyphs(t *testing.T) {
	var c cellbuffer
	c.init(3, 2)

	c.set(0, 0, '◤')
	c.set(2, 1, '▶')

	if got, want := c.String(), "◤  \n  ▶"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}
