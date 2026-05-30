package main

import (
	"math"
	"math/rand/v2"
)

type boid struct {
	pos           Point
	nextPos       Point
	vel           Point
	maxX, maxY    float64
	bounce        bool
	clampMinSpeed bool
}

func initBoidsOnScreenSize(cfg config, screenWidth, screenHeight int) []boid {
	count := screenWidth*screenHeight/effectiveCellsPerBoid(cfg) + 1
	return initRandomBoids(cfg, count, screenWidth, screenHeight)
}

func effectiveCellsPerBoid(cfg config) int {
	if cfg.cellsPerBoid <= 0 {
		return defaultCellsPerBoid
	}
	return cfg.cellsPerBoid
}

func initRandomBoids(cfg config, count int, screenWidth, screenHeight int) []boid {
	rng := newRandom(cfg)
	boids := make([]boid, count)
	for i := range boids {
		boids[i] = boid{
			pos: Point{
				x: rng.Float64() * float64(screenWidth),
				y: rng.Float64() * float64(screenHeight),
			},
			vel: Point{
				x: rng.Float64()*cfg.maxSpeed*2 - cfg.maxSpeed,
				y: rng.Float64()*cfg.maxSpeed*2 - cfg.maxSpeed,
			},
			maxX:          float64(screenWidth),
			maxY:          float64(screenHeight),
			bounce:        cfg.bounce,
			clampMinSpeed: cfg.clampMinSpeed,
		}
	}
	return boids
}

func newRandom(cfg config) *rand.Rand {
	if cfg.seed == nil {
		return rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	}
	return rand.New(rand.NewPCG(*cfg.seed, *cfg.seed^0x9e3779b97f4a7c15))
}

func (b *boid) update(boids []boid, cfg config) {
	accel := b.calcAcceleration(boids, cfg)
	b.applyAcceleration(accel, cfg)
}

func (b *boid) updateWithCandidateIndexes(boids []boid, candidateIndexes []int, cfg config) {
	accel := b.calcAccelerationWithCandidateIndexes(boids, candidateIndexes, cfg)
	b.applyAcceleration(accel, cfg)
}

func (b *boid) updateWithCandidateGrid(boids []boid, grid *spatialGrid, cfg config) {
	accel := b.calcAccelerationWithCandidateGrid(boids, grid, cfg)
	b.applyAcceleration(accel, cfg)
}

func (b *boid) applyAcceleration(accel Point, cfg config) {
	b.vel = b.vel.Add(accel).Limit(-cfg.maxSpeed, cfg.maxSpeed)

	if b.clampMinSpeed {
		if b.vel.x >= 0 && b.vel.x < cfg.targetMinSpeed {
			b.vel.x = Lerp(b.vel.x, cfg.targetMinSpeed, 0.5)
		}
		if b.vel.x < 0 && b.vel.x > -cfg.targetMinSpeed {
			b.vel.x = Lerp(b.vel.x, -cfg.targetMinSpeed, 0.5)
		}
		if b.vel.y >= 0 && b.vel.y < cfg.targetMinSpeed {
			b.vel.y = Lerp(b.vel.y, cfg.targetMinSpeed, 0.5)
		}
		if b.vel.y < 0 && b.vel.y > -cfg.targetMinSpeed {
			b.vel.y = Lerp(b.vel.y, -cfg.targetMinSpeed, 0.5)
		}
	}

	b.nextPos = b.pos.Add(b.vel)
	if !b.bounce {
		b.wrapAroundScreen()
	}
}

func (b *boid) wrapAroundScreen() {
	if b.nextPos.x < 0 {
		b.nextPos.x += b.maxX
	} else if b.nextPos.x > b.maxX {
		b.nextPos.x -= b.maxX
	}

	if b.nextPos.y < 0 {
		b.nextPos.y += b.maxY
	} else if b.nextPos.y > b.maxY {
		b.nextPos.y -= b.maxY
	}
}

func (b *boid) move() {
	b.pos = b.nextPos
}

func (b *boid) calcAcceleration(boids []boid, cfg config) Point {
	sep, avgPos, avgVel, count := b.measureNearby(boids, cfg)
	return b.calcAccelerationFromNearby(sep, avgPos, avgVel, count, cfg)
}

func (b *boid) calcAccelerationWithCandidateIndexes(boids []boid, candidateIndexes []int, cfg config) Point {
	sep, avgPos, avgVel, count := b.measureNearbyCandidateIndexes(boids, candidateIndexes, cfg)
	return b.calcAccelerationFromNearby(sep, avgPos, avgVel, count, cfg)
}

func (b *boid) calcAccelerationWithCandidateGrid(boids []boid, grid *spatialGrid, cfg config) Point {
	sep, avgPos, avgVel, count := b.measureNearbyCandidateGrid(boids, grid, cfg)
	return b.calcAccelerationFromNearby(sep, avgPos, avgVel, count, cfg)
}

func (b *boid) calcAccelerationFromNearby(sep, avgPos, avgVel Point, count int, cfg config) Point {
	accel := Point{}
	if b.bounce {
		accel = Point{bounce(b.pos.x, b.maxX, cfg), bounce(b.pos.y, b.maxY, cfg)}
	}

	if count == 0 {
		return accel
	}
	avgPos = avgPos.DivideV(float64(count))
	avgVel = avgVel.DivideV(float64(count))

	accelAlignment := avgVel.Subtract(b.vel).MultiplyV(cfg.adjustRate).MultiplyV(cfg.alignmentRate)
	accelCohesion := avgPos.Subtract(b.pos).MultiplyV(cfg.adjustRate).MultiplyV(cfg.cohesionRate)
	accelSeparation := sep.MultiplyV(cfg.adjustRate).MultiplyV(cfg.separationRate)

	accel = accel.Add(accelAlignment).Add(accelCohesion).Add(accelSeparation)
	return accel
}

func (b *boid) measureNearby(boids []boid, cfg config) (Point, Point, Point, int) {
	var sep, avgPos, avgVel Point
	count := 0
	if cfg.radius <= 0 {
		return sep, avgPos, avgVel, count
	}
	radiusSquared := cfg.radius * cfg.radius
	pos := b.pos

	for i := range boids {
		otherPos := boids[i].pos
		offset := pos.Subtract(otherPos)
		if offset.x == 0 && offset.y == 0 {
			continue
		}
		distanceSquared := offset.magnitudeSquared()
		if !(distanceSquared < radiusSquared) {
			continue
		}

		distance := math.Sqrt(distanceSquared)
		count++
		avgVel = avgVel.Add(boids[i].vel)
		avgPos = avgPos.Add(otherPos)
		sep = sep.Add(offset.DivideV(distance * 1.5))
	}
	return sep, avgPos, avgVel, count
}

func (b *boid) measureNearbyCandidateIndexes(boids []boid, candidateIndexes []int, cfg config) (Point, Point, Point, int) {
	var sep, avgPos, avgVel Point
	count := 0
	if cfg.radius <= 0 {
		return sep, avgPos, avgVel, count
	}
	radiusSquared := cfg.radius * cfg.radius
	pos := b.pos

	for _, candidateIndex := range candidateIndexes {
		otherPos := boids[candidateIndex].pos
		offset := pos.Subtract(otherPos)
		if offset.x == 0 && offset.y == 0 {
			continue
		}
		distanceSquared := offset.magnitudeSquared()
		if !(distanceSquared < radiusSquared) {
			continue
		}

		distance := math.Sqrt(distanceSquared)
		count++
		avgVel = avgVel.Add(boids[candidateIndex].vel)
		avgPos = avgPos.Add(otherPos)
		sep = sep.Add(offset.DivideV(distance * 1.5))
	}
	return sep, avgPos, avgVel, count
}

func (b *boid) measureNearbyCandidateGrid(boids []boid, grid *spatialGrid, cfg config) (Point, Point, Point, int) {
	var sep, avgPos, avgVel Point
	count := 0
	if cfg.radius <= 0 {
		return sep, avgPos, avgVel, count
	}
	radiusSquared := cfg.radius * cfg.radius
	pos := b.pos

	grid.visitNearbyBoidIndexes(pos, func(candidateIndex int) {
		otherPos := boids[candidateIndex].pos
		offset := pos.Subtract(otherPos)
		if offset.x == 0 && offset.y == 0 {
			return
		}
		distanceSquared := offset.magnitudeSquared()
		if !(distanceSquared < radiusSquared) {
			return
		}

		distance := math.Sqrt(distanceSquared)
		count++
		avgVel = avgVel.Add(boids[candidateIndex].vel)
		avgPos = avgPos.Add(otherPos)
		sep = sep.Add(offset.DivideV(distance * 1.5))
	})

	return sep, avgPos, avgVel, count
}

func bounce(pos, maxBorderPos float64, cfg config) float64 {
	if pos < cfg.radius {
		return 1 / pos
	} else if pos > maxBorderPos-cfg.radius {
		return 1 / (pos - maxBorderPos)
	}
	return 0
}
