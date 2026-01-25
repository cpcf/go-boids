package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	headlessMode := flag.Bool("headless", false, "run without Bubble Tea output")
	frames := flag.Int("frames", 1, "number of simulation steps in headless mode")
	width := flag.Int("width", 0, "simulation width in headless mode")
	height := flag.Int("height", 0, "simulation height in headless mode")
	flag.Parse()

	cfg := loadConfig()

	if *headlessMode {
		opts := headlessOptions{
			frames: *frames,
			width:  *width,
			height: *height,
		}
		summary, err := runHeadless(cfg, opts)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		if err := writeHeadlessSummary(os.Stdout, summary); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
		return
	}

	m := model{
		cfg: cfg,
		sim: newSimulation(cfg),
	}

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Println("Uh oh:", err)
		os.Exit(1)
	}
}
