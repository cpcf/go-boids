package main

const spatialGridBoidThreshold = 128

type simulation struct {
	cfg           config
	width, height int
	boids         []boid
	frames        int

	previousBoids []boid
	nearbyGrid    spatialGrid
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
}

func (s *simulation) SetConfig(cfg config) {
	s.cfg = cfg
	s.nearbyGrid.cellSize = cfg.radius
}

func (s *simulation) Step() {
	s.previousBoids = append(s.previousBoids[:0], s.boids...)

	if len(s.previousBoids) < spatialGridBoidThreshold {
		for i := range s.boids {
			s.boids[i].update(s.previousBoids, s.cfg)
		}
	} else {
		s.nearbyGrid.cellSize = s.cfg.radius
		s.nearbyGrid.rebuild(s.previousBoids)

		for i := range s.boids {
			s.boids[i].updateWithCandidateGrid(s.previousBoids, &s.nearbyGrid, s.cfg)
		}
	}

	for i := range s.boids {
		s.boids[i].move()
	}

	s.frames++
}

func (s *simulation) Boids() []boid {
	return s.boids
}
