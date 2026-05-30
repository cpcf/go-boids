package main

import "math"

type simulationStats struct {
	frames    int
	width     int
	height    int
	boids     int
	avgSpeed  float64
	minSpeed  float64
	maxSpeed  float64
	centroidX float64
	centroidY float64
}

func (s *simulation) Stats() simulationStats {
	stats := simulationStats{
		frames:    s.frames,
		width:     s.width,
		height:    s.height,
		boids:     len(s.x),
		avgSpeed:  0,
		minSpeed:  0,
		maxSpeed:  0,
		centroidX: 0,
		centroidY: 0,
	}

	if len(s.x) == 0 {
		return stats
	}

	totalSpeed := 0.0
	totalX := 0.0
	totalY := 0.0
	minSpeed := math.Inf(1)
	maxSpeed := 0.0

	for i := range s.x {
		speed := math.Hypot(s.vx[i], s.vy[i])
		totalSpeed += speed
		totalX += s.x[i]
		totalY += s.y[i]
		if speed < minSpeed {
			minSpeed = speed
		}
		if speed > maxSpeed {
			maxSpeed = speed
		}
	}

	stats.avgSpeed = totalSpeed / float64(len(s.x))
	stats.minSpeed = minSpeed
	stats.maxSpeed = maxSpeed
	stats.centroidX = totalX / float64(len(s.x))
	stats.centroidY = totalY / float64(len(s.x))

	return stats
}
