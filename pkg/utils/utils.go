package utils

import "math"

func TruncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}

func FatalError(err error) {
	if err != nil {
		panic(err)
	}
}
