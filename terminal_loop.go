package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	ansiEnterAltScreen = "\x1b[?1049h"
	ansiExitAltScreen  = "\x1b[?1049l"
	ansiHideCursor     = "\x1b[?25l"
	ansiShowCursor     = "\x1b[?25h"
	ansiClearScreen    = "\x1b[2J"
)

type parsedInput struct {
	quit bool
	key  rune
}

func parseTerminalKey(r rune) parsedInput {
	switch r {
	case 0x03: // Ctrl-C
		return parsedInput{quit: true}
	case 0x1b: // Esc
		return parsedInput{quit: true}
	default:
		return parsedInput{key: r}
	}
}

func applyParsedInput(m *model, input parsedInput) bool {
	if input.quit {
		return true
	}
	return m.handleInputRune(input.key)
}

func runInteractive(cfg config) error {
	return runInteractiveWithIO(cfg, os.Stdin, os.Stdout)
}

func runInteractiveWithIO(cfg config, input *os.File, output io.Writer) error {
	if input == nil {
		return fmt.Errorf("input is required")
	}
	if output == nil {
		return fmt.Errorf("output is required")
	}

	fd := int(input.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("interactive mode requires a terminal")
	}

	originalState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("enable raw terminal mode: %w", err)
	}
	defer func() { _ = term.Restore(fd, originalState) }()

	width, height, err := term.GetSize(fd)
	if err != nil {
		return fmt.Errorf("read terminal size: %w", err)
	}

	if _, err := fmt.Fprint(output, ansiEnterAltScreen+ansiHideCursor+ansiClearScreen); err != nil {
		return err
	}

	defer func() {
		_, _ = fmt.Fprint(output, ansiShowCursor+ansiExitAltScreen)
	}()

	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}
	m.handleWindowResize(width, height)

	renderer := sparseRenderer{}
	renderer.reset(m.cells.width(), m.cells.height())

	if err := renderer.renderSimulation(output, &m.sim, m.statusLine()); err != nil {
		return fmt.Errorf("initial render: %w", err)
	}

	reader := bufio.NewReader(input)
	keyCh := make(chan parsedInput, 16)
	resizeCh := make(chan os.Signal, 1)
	errorCh := make(chan error, 1)

	signal.Notify(resizeCh, syscall.SIGWINCH)
	defer signal.Stop(resizeCh)

	go func() {
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				errorCh <- err
				return
			}
			keyCh <- parseTerminalKey(r)
		}
	}()

	fps := cfg.fps
	if fps <= 0 {
		fps = defaultConfig().fps
	}
	frameInterval := time.Second / time.Duration(fps)
	ticker := time.NewTicker(frameInterval)
	defer ticker.Stop()

	renderFPS := effectiveRenderFPS(cfg)
	var renderTicker *time.Ticker
	var renderC <-chan time.Time
	if renderFPS < fps {
		renderTicker = time.NewTicker(time.Second / time.Duration(renderFPS))
		defer renderTicker.Stop()
		renderC = renderTicker.C
	}

	renderPending := false
	renderNow := func() error {
		renderPending = false
		return renderer.renderSimulation(output, &m.sim, m.statusLine())
	}

	for {
		select {
		case key := <-keyCh:
			if applyParsedInput(&m, key) {
				return nil
			}
			if err := renderNow(); err != nil {
				return fmt.Errorf("frame render: %w", err)
			}

		case <-ticker.C:
			if !m.paused {
				m.sim.Step()
				renderPending = true
				if renderC == nil {
					if err := renderNow(); err != nil {
						return fmt.Errorf("frame render: %w", err)
					}
				}
			}

		case <-renderC:
			if !renderPending {
				continue
			}
			if err := renderNow(); err != nil {
				return fmt.Errorf("frame render: %w", err)
			}

		case <-resizeCh:
			width, height, err := term.GetSize(fd)
			if err != nil {
				return fmt.Errorf("read terminal size on resize: %w", err)
			}
			m.handleWindowResize(width, height)
			renderer.reset(m.cells.width(), m.cells.height())
			if _, err := fmt.Fprint(output, ansiClearScreen); err != nil {
				return err
			}
			if err := renderNow(); err != nil {
				return fmt.Errorf("frame render: %w", err)
			}

		case err := <-errorCh:
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

func effectiveRenderFPS(cfg config) int {
	if cfg.renderFps > 0 {
		return cfg.renderFps
	}
	if cfg.fps > 0 {
		return cfg.fps
	}
	return defaultConfig().renderFps
}
