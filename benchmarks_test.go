package main

import "testing"

var benchmarkSizes = []struct {
	name          string
	width, height int
}{
	{name: "120x40", width: 120, height: 40},
	{name: "180x60", width: 180, height: 60},
	{name: "240x80", width: 240, height: 80},
}

var (
	benchPointSink  Point
	benchCountSink  int
	benchStringSink string
	benchBoidSink   boid
)

func BenchmarkUpdateBoids(b *testing.B) {
	cfg := defaultConfig()
	cfg.radius = 7
	cfg.maxSpeed = 0.5
	cfg.adjustRate = 0.025
	cfg.alignmentRate = 1
	cfg.cohesionRate = 1
	cfg.separationRate = 1
	cfg.targetMinSpeed = 0.01

	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			sim := newSimulation(cfg)
			sim.Resize(size.width, size.height)
			sim.boids = deterministicBoids(size.width, size.height)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sim.Step()
			}
			b.StopTimer()

			boids := sim.Boids()
			if len(boids) > 0 {
				benchBoidSink = boids[0]
			}
		})
	}
}

func BenchmarkMeasureNearby(b *testing.B) {
	cfg := defaultConfig()
	cfg.radius = 7

	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			boids := deterministicBoids(size.width, size.height)
			subject := boids[len(boids)/2]

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sep, avgPos, avgVel, count := subject.measureNearby(boids, cfg)
				benchPointSink = sep.Add(avgPos).Add(avgVel)
				benchCountSink = count
			}
		})
	}
}

func BenchmarkMeasureNearbyCandidateIndexes(b *testing.B) {
	cfg := defaultConfig()
	cfg.radius = 7

	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			boids := deterministicBoids(size.width, size.height)
			subject := boids[len(boids)/2]
			grid := newSpatialGrid(cfg.radius)
			grid.rebuild(boids)
			candidates := grid.candidateIndexes(subject.pos, make([]int, 0, len(boids)))

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sep, avgPos, avgVel, count := subject.measureNearbyCandidateIndexes(boids, candidates, cfg)
				benchPointSink = sep.Add(avgPos).Add(avgVel)
				benchCountSink = count
			}
		})
	}
}

func BenchmarkCellbufferWipe(b *testing.B) {
	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			var c cellbuffer
			c.init(size.width, size.height)
			fillCellbuffer(&c)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c.wipe()
			}
			b.StopTimer()

			benchCountSink = len(c.cells)
		})
	}
}

func BenchmarkCellbufferString(b *testing.B) {
	for _, size := range benchmarkSizes {
		b.Run(size.name, func(b *testing.B) {
			var c cellbuffer
			c.init(size.width, size.height)
			fillCellbuffer(&c)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchStringSink = c.String()
			}
		})
	}
}

func deterministicBoids(width, height int) []boid {
	count := width*height/125 + 1
	boids := make([]boid, count)
	velocities := []Point{
		{x: 0.31, y: 0.17},
		{x: -0.22, y: 0.29},
		{x: 0.18, y: -0.33},
		{x: -0.27, y: -0.19},
	}

	for i := range boids {
		x := float64((i*37)%(width-2) + 1)
		y := float64((i*23)%(height-2) + 1)
		vel := velocities[i%len(velocities)]
		boids[i] = boid{
			pos:           Point{x: x, y: y},
			nextPos:       Point{x: x, y: y},
			vel:           vel,
			maxX:          float64(width),
			maxY:          float64(height),
			bounce:        false,
			clampMinSpeed: true,
		}
	}

	return boids
}

func fillCellbuffer(c *cellbuffer) {
	glyphs := []rune{' ', '.', '#', '*'}
	for i := range c.cells {
		c.cells[i] = glyphs[i%len(glyphs)]
	}
}
