package main

import "math"

const spatialGridBoidThreshold = 128

type simulation struct {
	cfg    config
	width  int
	height int

	// SoA simulation state. x/y/vx/vy are the authoritative current-frame
	// position and velocity buffers.
	x, y   []float64
	vx, vy []float64
	nextX  []float64
	nextY  []float64
	nextVX []float64
	nextVY []float64
	maxX   []float64
	maxY   []float64
	bounce []bool
	clamp  []bool

	// boids is a lazy compatibility view for tests and the Bubble Tea renderer.
	boids      []boid
	boidsDirty bool
	frames     int

	nearbyGrid spatialGrid
}

func newSimulation(cfg config) simulation {
	return simulation{
		cfg:        cfg,
		nearbyGrid: newSpatialGrid(cfg.radius, 0, 0),
	}
}

func (s *simulation) Resize(width, height int) {
	s.width = width
	s.height = height
	s.nearbyGrid.resize(width, height)
	s.frames = 0
	s.setBoids(initBoidsOnScreenSize(s.cfg, width, height))
}

func (s *simulation) SetConfig(cfg config) {
	s.cfg = cfg
	s.nearbyGrid.setCellSize(cfg.radius)
}

func (s *simulation) Step() {
	if len(s.x) == 0 {
		s.frames++
		return
	}

	if cap(s.nextX) < len(s.x) {
		s.resizeStateBuffers(len(s.x))
	}

	nextCount := len(s.x)
	s.nextX = s.nextX[:nextCount]
	s.nextY = s.nextY[:nextCount]
	s.nextVX = s.nextVX[:nextCount]
	s.nextVY = s.nextVY[:nextCount]
	if len(s.x) < spatialGridBoidThreshold {
		for i := range s.x {
			sep, avgPos, avgVel, count := s.measureNearby(i, s.cfg)
			s.applyAccelerationAndMove(i, sep, avgPos, avgVel, count, s.cfg)
		}
	} else {
		s.nearbyGrid.setCellSize(s.cfg.radius)
		s.nearbyGrid.rebuildFromPositions(s.x, s.y)

		for i := range s.x {
			sep, avgPos, avgVel, count := s.measureNearbyWithGrid(i, s.cfg)
			s.applyAccelerationAndMove(i, sep, avgPos, avgVel, count, s.cfg)
		}
	}

	s.x, s.nextX = s.nextX, s.x
	s.y, s.nextY = s.nextY, s.y
	s.vx, s.nextVX = s.nextVX, s.vx
	s.vy, s.nextVY = s.nextVY, s.vy
	s.boidsDirty = true
	s.frames++
}

func (s *simulation) Boids() []boid {
	if s.boidsDirty {
		s.materializeBoids()
	}
	return s.boids
}

func (s *simulation) resizeStateBuffers(count int) {
	if cap(s.x) < count {
		s.x = make([]float64, count)
	} else {
		s.x = s.x[:count]
	}
	if cap(s.y) < count {
		s.y = make([]float64, count)
	} else {
		s.y = s.y[:count]
	}
	if cap(s.vx) < count {
		s.vx = make([]float64, count)
	} else {
		s.vx = s.vx[:count]
	}
	if cap(s.vy) < count {
		s.vy = make([]float64, count)
	} else {
		s.vy = s.vy[:count]
	}
	if cap(s.nextX) < count {
		s.nextX = make([]float64, count)
	} else {
		s.nextX = s.nextX[:count]
	}
	if cap(s.nextY) < count {
		s.nextY = make([]float64, count)
	} else {
		s.nextY = s.nextY[:count]
	}
	if cap(s.nextVX) < count {
		s.nextVX = make([]float64, count)
	} else {
		s.nextVX = s.nextVX[:count]
	}
	if cap(s.nextVY) < count {
		s.nextVY = make([]float64, count)
	} else {
		s.nextVY = s.nextVY[:count]
	}
	if cap(s.maxX) < count {
		s.maxX = make([]float64, count)
	} else {
		s.maxX = s.maxX[:count]
	}
	if cap(s.maxY) < count {
		s.maxY = make([]float64, count)
	} else {
		s.maxY = s.maxY[:count]
	}
	if cap(s.bounce) < count {
		s.bounce = make([]bool, count)
	} else {
		s.bounce = s.bounce[:count]
	}
	if cap(s.clamp) < count {
		s.clamp = make([]bool, count)
	} else {
		s.clamp = s.clamp[:count]
	}
	if cap(s.boids) < count {
		s.boids = make([]boid, count)
	} else {
		s.boids = s.boids[:count]
	}
}

func (s *simulation) setBoids(boids []boid) {
	count := len(boids)
	s.resizeStateBuffers(count)

	for i := 0; i < count; i++ {
		src := boids[i]
		s.x[i] = src.pos.x
		s.y[i] = src.pos.y
		s.vx[i] = src.vel.x
		s.vy[i] = src.vel.y
		s.maxX[i] = src.maxX
		s.maxY[i] = src.maxY
		s.bounce[i] = src.bounce
		s.clamp[i] = src.clampMinSpeed

		nextPos := src.pos
		s.boids[i] = src
		s.boids[i].nextPos = nextPos
		s.nextX[i] = src.pos.x
		s.nextY[i] = src.pos.y
		s.nextVX[i] = src.vel.x
		s.nextVY[i] = src.vel.y
	}
	s.boidsDirty = false
}

func (s *simulation) materializeBoids() {
	count := len(s.x)
	if cap(s.boids) < count {
		s.boids = make([]boid, count)
	} else {
		s.boids = s.boids[:count]
	}

	for i := 0; i < count; i++ {
		pos := Point{x: s.x[i], y: s.y[i]}
		s.boids[i] = boid{
			pos:           pos,
			nextPos:       pos,
			vel:           Point{x: s.vx[i], y: s.vy[i]},
			maxX:          s.maxX[i],
			maxY:          s.maxY[i],
			bounce:        s.bounce[i],
			clampMinSpeed: s.clamp[i],
		}
	}
	s.boidsDirty = false
}

func (s *simulation) applyAccelerationAndMove(i int, sep, avgPos, avgVel Point, count int, cfg config) {
	pos := Point{x: s.x[i], y: s.y[i]}
	vel := Point{x: s.vx[i], y: s.vy[i]}
	maxX := s.maxX[i]
	maxY := s.maxY[i]

	accel := Point{}
	if s.bounce[i] {
		accel = Point{bounce(pos.x, maxX, cfg), bounce(pos.y, maxY, cfg)}
	}

	if count != 0 {
		avgPos = avgPos.DivideV(float64(count))
		avgVel = avgVel.DivideV(float64(count))

		accelAlignment := avgVel.Subtract(vel).MultiplyV(cfg.adjustRate).MultiplyV(cfg.alignmentRate)
		accelCohesion := avgPos.Subtract(pos).MultiplyV(cfg.adjustRate).MultiplyV(cfg.cohesionRate)
		accelSeparation := sep.MultiplyV(cfg.adjustRate).MultiplyV(cfg.separationRate)
		accel = accel.Add(accelAlignment).Add(accelCohesion).Add(accelSeparation)
	}

	vel = vel.Add(accel).Limit(-cfg.maxSpeed, cfg.maxSpeed)
	if s.clamp[i] {
		if vel.x >= 0 && vel.x < cfg.targetMinSpeed {
			vel.x = Lerp(vel.x, cfg.targetMinSpeed, 0.5)
		}
		if vel.x < 0 && vel.x > -cfg.targetMinSpeed {
			vel.x = Lerp(vel.x, -cfg.targetMinSpeed, 0.5)
		}
		if vel.y >= 0 && vel.y < cfg.targetMinSpeed {
			vel.y = Lerp(vel.y, cfg.targetMinSpeed, 0.5)
		}
		if vel.y < 0 && vel.y > -cfg.targetMinSpeed {
			vel.y = Lerp(vel.y, -cfg.targetMinSpeed, 0.5)
		}
	}

	nextPos := pos.Add(vel)
	if !s.bounce[i] {
		if nextPos.x < 0 {
			nextPos.x += maxX
		} else if nextPos.x > maxX {
			nextPos.x -= maxX
		}

		if nextPos.y < 0 {
			nextPos.y += maxY
		} else if nextPos.y > maxY {
			nextPos.y -= maxY
		}
	}

	s.nextX[i] = nextPos.x
	s.nextY[i] = nextPos.y
	s.nextVX[i] = vel.x
	s.nextVY[i] = vel.y
}

func (s *simulation) measureNearby(i int, cfg config) (Point, Point, Point, int) {
	var sep, avgPos, avgVel Point
	count := 0
	if cfg.radius <= 0 {
		return sep, avgPos, avgVel, count
	}

	radiusSquared := cfg.radius * cfg.radius
	posX := s.x[i]
	posY := s.y[i]

	for j := range s.x {
		if j == i {
			continue
		}

		offsetX := posX - s.x[j]
		offsetY := posY - s.y[j]
		if offsetX == 0 && offsetY == 0 {
			continue
		}

		distanceSquared := offsetX*offsetX + offsetY*offsetY
		if !(distanceSquared < radiusSquared) {
			continue
		}

		distance := math.Sqrt(distanceSquared)
		count++
		avgVel = avgVel.Add(Point{x: s.vx[j], y: s.vy[j]})
		avgPos = avgPos.Add(Point{x: s.x[j], y: s.y[j]})
		sep = sep.Add(Point{x: offsetX, y: offsetY}.DivideV(distance * 1.5))
	}

	return sep, avgPos, avgVel, count
}

func (s *simulation) measureNearbyWithGrid(i int, cfg config) (Point, Point, Point, int) {
	var sep, avgPos, avgVel Point
	count := 0
	if cfg.radius <= 0 {
		return sep, avgPos, avgVel, count
	}

	radiusSquared := cfg.radius * cfg.radius
	posX := s.x[i]
	posY := s.y[i]

	s.nearbyGrid.visitNearbyBoidIndexes(Point{x: posX, y: posY}, func(j int) {
		if j == i {
			return
		}

		offsetX := posX - s.x[j]
		offsetY := posY - s.y[j]
		if offsetX == 0 && offsetY == 0 {
			return
		}

		distanceSquared := offsetX*offsetX + offsetY*offsetY
		if !(distanceSquared < radiusSquared) {
			return
		}

		distance := math.Sqrt(distanceSquared)
		count++
		avgVel = avgVel.Add(Point{x: s.vx[j], y: s.vy[j]})
		avgPos = avgPos.Add(Point{x: s.x[j], y: s.y[j]})
		sep = sep.Add(Point{x: offsetX, y: offsetY}.DivideV(distance * 1.5))
	})

	return sep, avgPos, avgVel, count
}
