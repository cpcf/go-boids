package main

import (
	"math"
	"slices"
	"testing"
)

func TestInitBoidsOnScreenSizeCount(t *testing.T) {
	tests := []struct {
		name         string
		cellsPerBoid int
		want         int
	}{
		{
			name:         "default",
			cellsPerBoid: defaultCellsPerBoid,
			want:         14,
		},
		{
			name:         "configured",
			cellsPerBoid: 100,
			want:         11,
		},
		{
			name:         "invalid zero falls back to default",
			cellsPerBoid: 0,
			want:         14,
		},
		{
			name:         "invalid negative falls back to default",
			cellsPerBoid: -50,
			want:         14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			cfg.cellsPerBoid = tt.cellsPerBoid

			boids := initBoidsOnScreenSize(cfg, 40, 25)

			if got := len(boids); got != tt.want {
				t.Fatalf("len(initBoidsOnScreenSize()) = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestInitBoidsUsesCellsPerBoidFromEnvConfig(t *testing.T) {
	configWithCleanEnv(t)
	configWithDotenv(t, `CELLS_PER_BOID=100`)

	cfg := loadConfig()
	boids := initBoidsOnScreenSize(cfg, 40, 25)

	if got, want := len(boids), 11; got != want {
		t.Fatalf("len(initBoidsOnScreenSize()) = %d, want %d", got, want)
	}
}

func TestInitBoidsUsesProvidedConfig(t *testing.T) {
	cfg := defaultConfig()
	cfg.cellsPerBoid = 100
	cfg.maxSpeed = 0.1
	cfg.bounce = false
	cfg.clampMinSpeed = false

	boids := initBoidsOnScreenSize(cfg, 40, 25)
	if got, want := len(boids), 11; got != want {
		t.Fatalf("len(initBoidsOnScreenSize()) = %d, want %d", got, want)
	}

	for i := range boids {
		if boids[i].bounce != cfg.bounce {
			t.Fatalf("boid[%d].bounce = %v, want %v", i, boids[i].bounce, cfg.bounce)
		}
		if boids[i].clampMinSpeed != cfg.clampMinSpeed {
			t.Fatalf("boid[%d].clampMinSpeed = %v, want %v", i, boids[i].clampMinSpeed, cfg.clampMinSpeed)
		}
		if boids[i].vel.x < -cfg.maxSpeed || boids[i].vel.x > cfg.maxSpeed {
			t.Fatalf("boid[%d].vel.x = %f, expected within [-%f, %f]", i, boids[i].vel.x, cfg.maxSpeed, cfg.maxSpeed)
		}
		if boids[i].vel.y < -cfg.maxSpeed || boids[i].vel.y > cfg.maxSpeed {
			t.Fatalf("boid[%d].vel.y = %f, expected within [-%f, %f]", i, boids[i].vel.y, cfg.maxSpeed, cfg.maxSpeed)
		}
	}
}

func TestInitRandomBoidsDeterministicWithSeed(t *testing.T) {
	cfg := defaultConfig()
	cfg.seed = uint64Ptr(1234)

	gotA := initRandomBoids(cfg, 12, 80, 50)
	gotB := initRandomBoids(cfg, 12, 80, 50)

	if len(gotA) != len(gotB) {
		t.Fatalf("len mismatch: %d != %d", len(gotA), len(gotB))
	}

	for i := range gotA {
		if gotA[i] != gotB[i] {
			t.Fatalf("boid[%d] = %+v, want %+v", i, gotA[i], gotB[i])
		}
	}
}

func TestInitRandomBoidsDiffersWithDifferentSeed(t *testing.T) {
	cfgA := defaultConfig()
	cfgA.seed = uint64Ptr(1)
	cfgB := defaultConfig()
	cfgB.seed = uint64Ptr(2)

	boidsA := initRandomBoids(cfgA, 12, 80, 50)
	boidsB := initRandomBoids(cfgB, 12, 80, 50)

	differ := false
	for i := range boidsA {
		if boidsA[i] != boidsB[i] {
			differ = true
			break
		}
	}
	if !differ {
		t.Fatalf("expected initial boids to differ with different seeds")
	}
}

func uint64Ptr(v uint64) *uint64 {
	return &v
}

func TestBoidWrapAroundScreen(t *testing.T) {
	b := boid{
		nextPos: Point{x: -1, y: 11},
		maxX:    10,
		maxY:    10,
	}

	b.wrapAroundScreen()

	if got, want := b.nextPos, (Point{x: 9, y: 1}); got != want {
		t.Fatalf("nextPos = %+v, want %+v", got, want)
	}
}

func TestBoidMove(t *testing.T) {
	b := boid{
		pos:     Point{x: 1, y: 2},
		nextPos: Point{x: 3, y: 4},
	}

	b.move()

	if got, want := b.pos, (Point{x: 3, y: 4}); got != want {
		t.Fatalf("pos = %+v, want %+v", got, want)
	}
}

func TestBoidUpdateWraps(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 1
	cfg.maxSpeed = 2
	cfg.clampMinSpeed = false

	b := boid{
		pos:           Point{x: 9.5, y: 2},
		vel:           Point{x: 1, y: 0},
		maxX:          10,
		maxY:          5,
		bounce:        false,
		clampMinSpeed: false,
	}

	b.update([]boid{b}, cfg)

	if got, want := b.nextPos, (Point{x: 0.5, y: 2}); got != want {
		t.Fatalf("nextPos = %+v, want %+v", got, want)
	}
}

func TestBoidApplyAccelerationUsesConfigMaxSpeed(t *testing.T) {
	cfg := defaultConfig()
	cfg.maxSpeed = 1
	cfg.clampMinSpeed = false

	b := boid{
		vel:           Point{x: 0.8, y: 0.9},
		clampMinSpeed: false,
		maxX:          10,
		maxY:          10,
	}

	b.applyAcceleration(Point{x: 1, y: 2}, cfg)

	if got, want := b.vel, (Point{x: 1, y: 1}); got != want {
		t.Fatalf("vel = %+v, want %+v", got, want)
	}
}

func TestMeasureNearby(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 5

	subject := boid{
		pos: Point{x: 10, y: 10},
		vel: Point{x: 1, y: 0},
	}
	boids := []boid{
		subject,
		{pos: Point{x: 13, y: 10}, vel: Point{x: 2, y: 1}},
		{pos: Point{x: 10, y: 14}, vel: Point{x: 0, y: 3}},
		{pos: Point{x: 30, y: 30}, vel: Point{x: 7, y: 7}},
	}

	sep, avgPos, avgVel, count := subject.measureNearby(boids, cfg)

	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if want := (Point{x: -2.0 / 3.0, y: -2.0 / 3.0}); !pointNear(sep, want) {
		t.Fatalf("sep = %+v, want %+v", sep, want)
	}
	if want := (Point{x: 23, y: 24}); avgPos != want {
		t.Fatalf("avgPos = %+v, want %+v", avgPos, want)
	}
	if want := (Point{x: 2, y: 4}); avgVel != want {
		t.Fatalf("avgVel = %+v, want %+v", avgVel, want)
	}
}

func TestMeasureNearbyCandidateIndexesMatchesFullScan(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 5
	boids := candidateGridTestBoids()
	grid := newSpatialGrid(cfg.radius, 80, 80)
	grid.rebuild(boids)

	tests := []struct {
		name          string
		subjectIndex  int
		wantCandidate int
		wantDistant   int
	}{
		{
			name:          "near upper boundaries",
			subjectIndex:  0,
			wantCandidate: 4,
			wantDistant:   6,
		},
		{
			name:          "exactly on boundaries",
			subjectIndex:  7,
			wantCandidate: 4,
			wantDistant:   6,
		},
		{
			name:          "lower boundary cells",
			subjectIndex:  10,
			wantCandidate: 13,
			wantDistant:   14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subject := boids[tt.subjectIndex]
			candidates := grid.candidateIndexes(subject.pos, nil)

			if !slices.Contains(candidates, tt.wantCandidate) {
				t.Fatalf("test setup expected candidate indexes to include outside-radius boid %d, got %v", tt.wantCandidate, candidates)
			}
			if slices.Contains(candidates, tt.wantDistant) {
				t.Fatalf("test setup expected candidate indexes to exclude distant boid %d, got %v", tt.wantDistant, candidates)
			}
			assertCandidatesIncludeNearby(t, subject, boids, candidates, cfg)
			assertNearbyMeasurementsMatch(t, subject, boids, candidates, cfg)
		})
	}
}

func TestMeasureNearbyCandidateGridMatchesFullScan(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 5

	boids := candidateGridTestBoids()
	grid := newSpatialGrid(cfg.radius, 80, 80)
	grid.rebuild(boids)

	for _, subjectIndex := range []int{0, 7, 10} {
		subject := boids[subjectIndex]
		wantSep, wantAvgPos, wantAvgVel, wantCount := subject.measureNearby(boids, cfg)
		gotSep, gotAvgPos, gotAvgVel, gotCount := subject.measureNearbyCandidateGrid(boids, &grid, cfg)

		if gotCount != wantCount {
			t.Fatalf("boid %d count = %d, want %d", subjectIndex, gotCount, wantCount)
		}
		if !pointNear(gotSep, wantSep) {
			t.Fatalf("boid %d sep = %+v, want %+v", subjectIndex, gotSep, wantSep)
		}
		if !pointNear(gotAvgPos, wantAvgPos) {
			t.Fatalf("boid %d avgPos = %+v, want %+v", subjectIndex, gotAvgPos, wantAvgPos)
		}
		if !pointNear(gotAvgVel, wantAvgVel) {
			t.Fatalf("boid %d avgVel = %+v, want %+v", subjectIndex, gotAvgVel, wantAvgVel)
		}
	}
}

func TestMeasureNearbyCandidateGridWithInvalidRadius(t *testing.T) {
	boids := candidateGridTestBoids()

	tests := []struct {
		name   string
		radius float64
		want   int
	}{
		{name: "zero", radius: 0, want: 0},
		{name: "negative", radius: -3, want: 0},
		{name: "nan", radius: math.NaN(), want: 0},
		{name: "infinity", radius: math.Inf(1), want: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			cfg.radius = tt.radius
			grid := newSpatialGrid(tt.radius, 80, 80)
			grid.rebuild(boids)

			for _, subject := range boids {
				want := tt.want
				if want < 0 {
					candidateIndexes := grid.candidateIndexes(subject.pos, nil)
					_, _, _, want = subject.measureNearbyCandidateIndexes(boids, candidateIndexes, cfg)
				}

				_, _, _, count := subject.measureNearbyCandidateGrid(boids, &grid, cfg)
				if got := count; got != want {
					t.Fatalf("radius=%v count = %d, want %d", tt.radius, got, want)
				}
			}
		})
	}
}

func candidateGridTestBoids() []boid {
	const offset = 20.0

	return []boid{
		{pos: Point{x: 9.9 + offset, y: 9.9 + offset}, vel: Point{x: 1, y: 0}},    // near upper cell boundaries
		{pos: Point{x: 14.8 + offset, y: 9.9 + offset}, vel: Point{x: 2, y: 1}},   // inside radius, adjacent x cell
		{pos: Point{x: 9.9 + offset, y: 10.2 + offset}, vel: Point{x: 0, y: 3}},   // inside radius, adjacent y cell
		{pos: Point{x: 13.3 + offset, y: 13.3 + offset}, vel: Point{x: -1, y: 2}}, // inside radius, diagonal cell
		{pos: Point{x: 14.95 + offset, y: 9.9 + offset}, vel: Point{x: 9, y: 9}},  // candidate, outside radius
		{pos: Point{x: 9.9 + offset, y: 9.9 + offset}, vel: Point{x: 4, y: 4}},    // same position excluded
		{pos: Point{x: 30 + offset, y: 30 + offset}, vel: Point{x: 7, y: 7}},      // outside candidate cells
		{pos: Point{x: 10 + offset, y: 10 + offset}, vel: Point{x: -2, y: -2}},    // exactly on x/y cell boundaries
		{pos: Point{x: 5.05 + offset, y: 10 + offset}, vel: Point{x: 1, y: -1}},   // inside radius across lower x boundary
		{pos: Point{x: 10 + offset, y: 14.95 + offset}, vel: Point{x: 2, y: -3}},  // inside radius across upper y boundary
		{pos: Point{x: 19.9, y: 19.9}, vel: Point{x: -3, y: 1}},                   // boundary probe
		{pos: Point{x: 24.8, y: 19.9}, vel: Point{x: 3, y: -1}},                   // inside radius across lower x boundary
		{pos: Point{x: 19.9, y: 14.95}, vel: Point{x: -1, y: -2}},                 // inside radius across upper y boundary
		{pos: Point{x: 14.8, y: 19.9}, vel: Point{x: -2, y: 0.5}},                 // candidate, outside radius
		{pos: Point{x: 0, y: 0}, vel: Point{x: -7, y: -7}},                        // outside candidate cells
	}
}

func assertCandidatesIncludeNearby(t *testing.T, subject boid, boids []boid, candidates []int, cfg config) {
	t.Helper()

	radiusSquared := cfg.radius * cfg.radius
	for i := range boids {
		if subject.pos.distanceSquared(boids[i].pos) < radiusSquared && !slices.Contains(candidates, i) {
			t.Fatalf("candidate indexes = %v, missing nearby boid %d at %+v for subject at %+v", candidates, i, boids[i].pos, subject.pos)
		}
	}
}

func assertNearbyMeasurementsMatch(t *testing.T, subject boid, boids []boid, candidates []int, cfg config) {
	t.Helper()

	wantSep, wantAvgPos, wantAvgVel, wantCount := subject.measureNearby(boids, cfg)
	gotSep, gotAvgPos, gotAvgVel, gotCount := subject.measureNearbyCandidateIndexes(boids, candidates, cfg)

	if gotCount != wantCount {
		t.Fatalf("count = %d, want %d", gotCount, wantCount)
	}
	if !pointNear(gotSep, wantSep) {
		t.Fatalf("sep = %+v, want %+v", gotSep, wantSep)
	}
	if !pointNear(gotAvgPos, wantAvgPos) {
		t.Fatalf("avgPos = %+v, want %+v", gotAvgPos, wantAvgPos)
	}
	if !pointNear(gotAvgVel, wantAvgVel) {
		t.Fatalf("avgVel = %+v, want %+v", gotAvgVel, wantAvgVel)
	}
}

func TestMeasureNearbyUsesConfigRadius(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 2

	subject := boid{pos: Point{x: 1, y: 1}}
	boids := []boid{
		subject,
		{pos: Point{x: 2.5, y: 1}}, // inside radius
		{pos: Point{x: 4.5, y: 1}}, // outside radius 2
	}

	_, _, _, count := subject.measureNearby(boids, cfg)
	if got, want := count, 1; got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
}

func TestBounce(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 7

	tests := []struct {
		name         string
		pos          float64
		maxBorderPos float64
		want         float64
	}{
		{
			name:         "inside low border",
			pos:          2,
			maxBorderPos: 20,
			want:         0.5,
		},
		{
			name:         "inside high border",
			pos:          19,
			maxBorderPos: 20,
			want:         -1,
		},
		{
			name:         "away from borders",
			pos:          10,
			maxBorderPos: 20,
			want:         0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := bounce(tt.pos, tt.maxBorderPos, cfg); !near(got, tt.want) {
				t.Fatalf("bounce() = %v, want %v", got, tt.want)
			}
		})
	}
}
