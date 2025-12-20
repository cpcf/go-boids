package main

const spatialGridBoidThreshold = 128

type simulation struct {
	cfg              config
	previousBoids    []boid
	nearbyGrid       spatialGrid
	candidateIndexes []int
}

func newSimulation(cfg config) simulation {
	return simulation{
		cfg:        cfg,
		nearbyGrid: newSpatialGrid(cfg.radius),
	}
}

func (s *simulation) Step(boids []boid) {
	s.previousBoids = append(s.previousBoids[:0], boids...)

	if len(s.previousBoids) < spatialGridBoidThreshold {
		for i := range boids {
			boids[i].update(s.previousBoids, s.cfg)
		}
	} else {
		s.nearbyGrid.cellSize = s.cfg.radius
		s.nearbyGrid.rebuild(s.previousBoids)
		if cap(s.candidateIndexes) < len(s.previousBoids) {
			s.candidateIndexes = make([]int, 0, len(s.previousBoids))
		}

		for i := range boids {
			s.candidateIndexes = s.nearbyGrid.candidateIndexes(s.previousBoids[i].pos, s.candidateIndexes)
			boids[i].updateWithCandidateIndexes(s.previousBoids, s.candidateIndexes, s.cfg)
		}
	}

	for i := range boids {
		boids[i].move()
	}
}
