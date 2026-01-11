package terma

import "math"

// EasingFunc defines the easing curve for an animation.
// Takes t in range [0, 1] and returns the eased value in range [0, 1].
type EasingFunc func(t float64) float64

// EaseLinear provides no easing (constant velocity).
func EaseLinear(t float64) float64 {
	return t
}

// EaseInQuad starts slow and accelerates (quadratic).
func EaseInQuad(t float64) float64 {
	return t * t
}

// EaseOutQuad starts fast and decelerates (quadratic).
func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

// EaseInOutQuad accelerates then decelerates (quadratic).
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// EaseInCubic starts slow and accelerates (cubic).
func EaseInCubic(t float64) float64 {
	return t * t * t
}

// EaseOutCubic starts fast and decelerates (cubic).
func EaseOutCubic(t float64) float64 {
	t--
	return t*t*t + 1
}

// EaseInOutCubic accelerates then decelerates (cubic).
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return (t-1)*(2*t-2)*(2*t-2) + 1
}

// EaseInQuart starts slow and accelerates (quartic).
func EaseInQuart(t float64) float64 {
	return t * t * t * t
}

// EaseOutQuart starts fast and decelerates (quartic).
func EaseOutQuart(t float64) float64 {
	t--
	return 1 - t*t*t*t
}

// EaseInOutQuart accelerates then decelerates (quartic).
func EaseInOutQuart(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	t--
	return 1 - 8*t*t*t*t
}

// EaseInQuint starts slow and accelerates (quintic).
func EaseInQuint(t float64) float64 {
	return t * t * t * t * t
}

// EaseOutQuint starts fast and decelerates (quintic).
func EaseOutQuint(t float64) float64 {
	t--
	return 1 + t*t*t*t*t
}

// EaseInOutQuint accelerates then decelerates (quintic).
func EaseInOutQuint(t float64) float64 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	}
	t--
	return 1 + 16*t*t*t*t*t
}

// EaseInSine starts slow using sine curve.
func EaseInSine(t float64) float64 {
	return 1 - math.Cos(t*math.Pi/2)
}

// EaseOutSine decelerates using sine curve.
func EaseOutSine(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

// EaseInOutSine accelerates then decelerates using sine curve.
func EaseInOutSine(t float64) float64 {
	return -(math.Cos(math.Pi*t) - 1) / 2
}

// EaseInExpo starts very slow with exponential acceleration.
func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

// EaseOutExpo decelerates exponentially.
func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

// EaseInOutExpo accelerates then decelerates exponentially.
func EaseInOutExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return math.Pow(2, 20*t-10) / 2
	}
	return (2 - math.Pow(2, -20*t+10)) / 2
}

// EaseInCirc starts slow with circular motion.
func EaseInCirc(t float64) float64 {
	return 1 - math.Sqrt(1-t*t)
}

// EaseOutCirc decelerates with circular motion.
func EaseOutCirc(t float64) float64 {
	t--
	return math.Sqrt(1 - t*t)
}

// EaseInOutCirc accelerates then decelerates with circular motion.
func EaseInOutCirc(t float64) float64 {
	if t < 0.5 {
		return (1 - math.Sqrt(1-4*t*t)) / 2
	}
	return (math.Sqrt(1-math.Pow(-2*t+2, 2)) + 1) / 2
}

// EaseInElastic provides elastic effect at the start.
func EaseInElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*5*math.Pi)
}

// EaseOutElastic provides elastic effect at the end.
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t)*math.Sin((t-0.1)*5*math.Pi) + 1
}

// EaseInOutElastic provides elastic effect at both ends.
func EaseInOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	t *= 2
	if t < 1 {
		return -0.5 * math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*5*math.Pi)
	}
	return 0.5*math.Pow(2, -10*(t-1))*math.Sin((t-1.1)*5*math.Pi) + 1
}

// EaseInBack overshoots slightly at the start.
func EaseInBack(t float64) float64 {
	const c1 = 1.70158
	const c3 = c1 + 1
	return c3*t*t*t - c1*t*t
}

// EaseOutBack overshoots slightly at the end.
func EaseOutBack(t float64) float64 {
	const c1 = 1.70158
	const c3 = c1 + 1
	t--
	return 1 + c3*t*t*t + c1*t*t
}

// EaseInOutBack overshoots slightly at both ends.
func EaseInOutBack(t float64) float64 {
	const c1 = 1.70158
	const c2 = c1 * 1.525
	if t < 0.5 {
		return (math.Pow(2*t, 2) * ((c2+1)*2*t - c2)) / 2
	}
	return (math.Pow(2*t-2, 2)*((c2+1)*(2*t-2)+c2) + 2) / 2
}

// EaseInBounce bounces at the start.
func EaseInBounce(t float64) float64 {
	return 1 - EaseOutBounce(1-t)
}

// EaseOutBounce bounces at the end.
func EaseOutBounce(t float64) float64 {
	const n1 = 7.5625
	const d1 = 2.75

	if t < 1/d1 {
		return n1 * t * t
	}
	if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	}
	if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

// EaseInOutBounce bounces at both ends.
func EaseInOutBounce(t float64) float64 {
	if t < 0.5 {
		return (1 - EaseOutBounce(1-2*t)) / 2
	}
	return (1 + EaseOutBounce(2*t-1)) / 2
}
