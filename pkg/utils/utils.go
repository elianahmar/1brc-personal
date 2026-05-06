package utils

import (
	"fmt"
	"math"
)

func TruncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}

func PanicOnError[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func PanicOnCondition(cond bool, msg string) {
	if cond {
		panic(fmt.Errorf(msg))
	}
}
