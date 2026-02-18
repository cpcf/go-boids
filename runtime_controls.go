package main

import "math"

const runtimeRadiusStep = 0.5
const runtimeMaxSpeedStep = 0.1
const minRuntimeRadius = 0.5
const minRuntimeMaxSpeed = 0.1

type runtimeAdjustment struct {
	radiusStep   float64
	maxSpeedStep float64
}

func runtimeAdjustmentForKey(r rune) (runtimeAdjustment, bool) {
	switch r {
	case '[':
		return runtimeAdjustment{radiusStep: -runtimeRadiusStep}, true
	case ']':
		return runtimeAdjustment{radiusStep: runtimeRadiusStep}, true
	case '-':
		return runtimeAdjustment{maxSpeedStep: -runtimeMaxSpeedStep}, true
	case '=':
		return runtimeAdjustment{maxSpeedStep: runtimeMaxSpeedStep}, true
	default:
		return runtimeAdjustment{}, false
	}
}

func applyRuntimeAdjustment(cfg config, adjustment runtimeAdjustment) config {
	cfg.radius = math.Max(minRuntimeRadius, cfg.radius+adjustment.radiusStep)
	cfg.maxSpeed = math.Max(minRuntimeMaxSpeed, cfg.maxSpeed+adjustment.maxSpeedStep)
	return cfg
}
