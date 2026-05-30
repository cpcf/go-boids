package main

const spatialGridBoidThreshold = 128

type simulation struct {
	cfg           config
	width, height int
	boids         []boid
	nextBoids     []boid
	frames        int

	nearbyGrid spatialGrid
}

func newSimulation(cfg config) simulation {
	return simulation{
		cfg:        cfg,
		nearbyGrid: newSpatialGrid(cfg.radius),
	}
}

func (s *simulation) Resize(width, height int) {
	s.width = width
	s.height = height
	s.frames = 0
	s.boids = initBoidsOnScreenSize(s.cfg, width, height)
	// Keep a reusable destination buffer with matching capacity for the next frame.
	s.nextBoids = make([]boid, 0, len(s.boids))
}

func (s *simulation) SetConfig(cfg config) {
	s.cfg = cfg
	s.nearbyGrid.cellSize = cfg.radius
}

func (s *simulation) Step() {
	if cap(s.nextBoids) < len(s.boids) {
		s.nextBoids = make([]boid, len(s.boids))
	} else {
		s.nextBoids = s.nextBoids[:len(s.boids)]
	}

	currentBoids := s.boids
	nextBoids := s.nextBoids

	if len(currentBoids) < spatialGridBoidThreshold {
		for i := range s.boids {
			nextBoids[i] = currentBoids[i]
			nextBoids[i].update(currentBoids, s.cfg)
		}
	} else {
		s.nearbyGrid.cellSize = s.cfg.radius
		s.nearbyGrid.rebuild(currentBoids)

		for i := range s.boids {
			nextBoids[i] = currentBoids[i]
			nextBoids[i].updateWithCandidateGrid(currentBoids, &s.nearbyGrid, s.cfg)
		}
	}

	for i := range nextBoids {
		nextBoids[i].move()
	}

	s.boids, s.nextBoids = nextBoids, currentBoids
	s.frames++
}

func (s *simulation) Boids() []boid {
	return s.boids
}
