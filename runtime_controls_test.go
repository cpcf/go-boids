package main

import (
	"math"
	"testing"
)

func TestRuntimeAdjustmentForKey(t *testing.T) {
	tests := []struct {
		name       string
		key        rune
		want       runtimeAdjustment
		shouldFind bool
	}{
		{
			name:       "decrease radius",
			key:        '[',
			want:       runtimeAdjustment{radiusStep: -runtimeRadiusStep},
			shouldFind: true,
		},
		{
			name:       "increase radius",
			key:        ']',
			want:       runtimeAdjustment{radiusStep: runtimeRadiusStep},
			shouldFind: true,
		},
		{
			name:       "decrease max speed",
			key:        '-',
			want:       runtimeAdjustment{maxSpeedStep: -runtimeMaxSpeedStep},
			shouldFind: true,
		},
		{
			name:       "increase max speed",
			key:        '=',
			want:       runtimeAdjustment{maxSpeedStep: runtimeMaxSpeedStep},
			shouldFind: true,
		},
		{
			name:       "unknown key",
			key:        'x',
			want:       runtimeAdjustment{},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := runtimeAdjustmentForKey(tt.key)
			if ok != tt.shouldFind {
				t.Fatalf("runtimeAdjustmentForKey(%q) found=%v, want %v", tt.key, ok, tt.shouldFind)
			}
			if got != tt.want {
				t.Fatalf("runtimeAdjustmentForKey(%q) adjustment=%+v, want %+v", tt.key, got, tt.want)
			}
		})
	}
}

func TestApplyRuntimeAdjustmentRadiusStepAndClamp(t *testing.T) {
	cfg := defaultConfig()
	cfg.radius = 1.0

	cfg = applyRuntimeAdjustment(cfg, runtimeAdjustment{radiusStep: -runtimeRadiusStep})
	if !almostEqualFloatForTest(cfg.radius, 0.5) {
		t.Fatalf("radius after one decrease = %.6f, want 0.500000", cfg.radius)
	}

	cfg = applyRuntimeAdjustment(cfg, runtimeAdjustment{radiusStep: -runtimeRadiusStep})
	if !almostEqualFloatForTest(cfg.radius, minRuntimeRadius) {
		t.Fatalf("radius after clamp = %.6f, want %.6f", cfg.radius, minRuntimeRadius)
	}
}

func TestApplyRuntimeAdjustmentMaxSpeedStepAndClamp(t *testing.T) {
	cfg := defaultConfig()
	cfg.maxSpeed = 0.2

	cfg = applyRuntimeAdjustment(cfg, runtimeAdjustment{maxSpeedStep: -runtimeMaxSpeedStep})
	if !almostEqualFloatForTest(cfg.maxSpeed, 0.1) {
		t.Fatalf("maxSpeed after one decrease = %.6f, want 0.100000", cfg.maxSpeed)
	}

	cfg = applyRuntimeAdjustment(cfg, runtimeAdjustment{maxSpeedStep: -runtimeMaxSpeedStep})
	if !almostEqualFloatForTest(cfg.maxSpeed, minRuntimeMaxSpeed) {
		t.Fatalf("maxSpeed after clamp = %.6f, want %.6f", cfg.maxSpeed, minRuntimeMaxSpeed)
	}
}

func almostEqualFloatForTest(a, b float64) bool {
	return math.Abs(a-b) <= 1e-9
}
