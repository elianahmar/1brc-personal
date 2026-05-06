package pkg

import "math"

func truncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}
