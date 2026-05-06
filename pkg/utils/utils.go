package utils

import "math"

func TruncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}

func PanicOnError[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}
