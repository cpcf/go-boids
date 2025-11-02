package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfigReturnsExpectedValues(t *testing.T) {
	got := defaultConfig()

	want := config{
		fps:            120,
		bounce:         true,
		clampMinSpeed:  true,
		cellsPerBoid:   defaultCellsPerBoid,
		radius:         7.0,
		maxSpeed:       1.0,
		adjustRate:     0.025,
		alignmentRate:  1.0,
		cohesionRate:   1.0,
		separationRate: 1.0,
		targetMinSpeed: 0.05,
	}

	if got != want {
		t.Fatalf("defaultConfig() = %+v, want %+v", got, want)
	}
}

func TestLoadConfigDefaultsWithoutDotenv(t *testing.T) {
	configWithCleanEnv(t)
	configWithNoDotenv(t)

	t.Setenv("FPS", "999")
	t.Setenv("BOUNCE", "false")
	t.Setenv("CLAMP_MIN_SPEED", "false")
	t.Setenv("CELLS_PER_BOID", "42")
	t.Setenv("RADIUS", "42.5")
	t.Setenv("MAX_SPEED", "9.9")
	t.Setenv("ADJUST_RATE", "0.9")
	t.Setenv("ALIGNMENT_RATE", "0.8")
	t.Setenv("COHESION_RATE", "0.7")
	t.Setenv("SEPARATION_RATE", "0.6")
	t.Setenv("TARGET_MIN_SPEED", "0.5")

	got := loadConfig()
	want := defaultConfig()

	if got != want {
		t.Fatalf("loadConfig() = %+v, want %+v", got, want)
	}
}

func TestLoadConfigReadsValuesFromDotenv(t *testing.T) {
	configWithCleanEnv(t)
	configWithDotenv(t, `FPS=240
BOUNCE=false
CLAMP_MIN_SPEED=false
CELLS_PER_BOID=64
RADIUS=9.5
MAX_SPEED=2.25
ADJUST_RATE=0.1
ALIGNMENT_RATE=1.5
COHESION_RATE=2
SEPARATION_RATE=3
TARGET_MIN_SPEED=0.25`)

	got := loadConfig()

	if got.fps != 240 {
		t.Fatalf("fps = %d, want 240", got.fps)
	}
	if got.bounce != false {
		t.Fatalf("bounce = %v, want false", got.bounce)
	}
	if got.clampMinSpeed != false {
		t.Fatalf("clampMinSpeed = %v, want false", got.clampMinSpeed)
	}
	if got.cellsPerBoid != 64 {
		t.Fatalf("cellsPerBoid = %d, want 64", got.cellsPerBoid)
	}
	if got.radius != 9.5 {
		t.Fatalf("radius = %f, want 9.5", got.radius)
	}
	if got.maxSpeed != 2.25 {
		t.Fatalf("maxSpeed = %f, want 2.25", got.maxSpeed)
	}
	if got.adjustRate != 0.1 {
		t.Fatalf("adjustRate = %f, want 0.1", got.adjustRate)
	}
	if got.alignmentRate != 1.5 {
		t.Fatalf("alignmentRate = %f, want 1.5", got.alignmentRate)
	}
	if got.cohesionRate != 2 {
		t.Fatalf("cohesionRate = %f, want 2", got.cohesionRate)
	}
	if got.separationRate != 3 {
		t.Fatalf("separationRate = %f, want 3", got.separationRate)
	}
	if got.targetMinSpeed != 0.25 {
		t.Fatalf("targetMinSpeed = %f, want 0.25", got.targetMinSpeed)
	}
}

func TestLoadConfigIgnoresEmptyEnvValues(t *testing.T) {
	configWithCleanEnv(t)
	configWithDotenv(t, `FPS=
BOUNCE=
CELLS_PER_BOID=`)

	t.Setenv("FPS", "777")
	t.Setenv("BOUNCE", "false")
	t.Setenv("CELLS_PER_BOID", "42")

	got := loadConfig()
	want := defaultConfig()

	if got.fps != want.fps {
		t.Fatalf("fps = %d, want %d", got.fps, want.fps)
	}
	if got.bounce != want.bounce {
		t.Fatalf("bounce = %v, want %v", got.bounce, want.bounce)
	}
	if got.cellsPerBoid != want.cellsPerBoid {
		t.Fatalf("cellsPerBoid = %d, want %d", got.cellsPerBoid, want.cellsPerBoid)
	}
}

func TestLoadConfigIgnoresParseErrors(t *testing.T) {
	configWithCleanEnv(t)
	configWithDotenv(t, `FPS=bad-int
BOUNCE=bad-bool
CLAMP_MIN_SPEED=bad-bool
CELLS_PER_BOID=bad-int
RADIUS=bad-float
MAX_SPEED=bad-float
ADJUST_RATE=bad-float
ALIGNMENT_RATE=bad-float
COHESION_RATE=bad-float
SEPARATION_RATE=bad-float
TARGET_MIN_SPEED=bad-float`)

	got := loadConfig()

	if got.fps != 0 {
		t.Fatalf("fps = %d, want 0", got.fps)
	}
	if got.bounce != false {
		t.Fatalf("bounce = %v, want false", got.bounce)
	}
	if got.clampMinSpeed != false {
		t.Fatalf("clampMinSpeed = %v, want false", got.clampMinSpeed)
	}
	if got.cellsPerBoid != defaultCellsPerBoid {
		t.Fatalf("cellsPerBoid = %d, want %d", got.cellsPerBoid, defaultCellsPerBoid)
	}
	if got.radius != 0 {
		t.Fatalf("radius = %f, want 0", got.radius)
	}
	if got.maxSpeed != 0 {
		t.Fatalf("maxSpeed = %f, want 0", got.maxSpeed)
	}
	if got.adjustRate != 0 {
		t.Fatalf("adjustRate = %f, want 0", got.adjustRate)
	}
	if got.alignmentRate != 0 {
		t.Fatalf("alignmentRate = %f, want 0", got.alignmentRate)
	}
	if got.cohesionRate != 0 {
		t.Fatalf("cohesionRate = %f, want 0", got.cohesionRate)
	}
	if got.separationRate != 0 {
		t.Fatalf("separationRate = %f, want 0", got.separationRate)
	}
	if got.targetMinSpeed != 0 {
		t.Fatalf("targetMinSpeed = %f, want 0", got.targetMinSpeed)
	}
}

func configWithNoDotenv(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func configWithDotenv(t *testing.T, contents string) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(contents), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Errorf("restore working directory: %v", err)
		}
	})
}

func configWithCleanEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"FPS",
		"BOUNCE",
		"CLAMP_MIN_SPEED",
		"CELLS_PER_BOID",
		"RADIUS",
		"MAX_SPEED",
		"ADJUST_RATE",
		"ALIGNMENT_RATE",
		"COHESION_RATE",
		"SEPARATION_RATE",
		"TARGET_MIN_SPEED",
	} {
		t.Setenv(key, "")
	}
}
