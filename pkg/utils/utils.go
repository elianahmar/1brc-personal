package utils

import (
	"fmt"
	"math"
	"time"
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
		panic(msg)
	}
}

// First is a simple util to find first occurrence of something based
// on condition and return the index. Using this primarily for reconciling the chunks
func First[T any](items []T, fn func(T) bool) int {
	for i := range items {
		if fn(items[i]) {
			return i
		}
	}
	return -1
}

func DayMonthYear() string {
	timestamp := time.Now()
	return fmt.Sprintf("%d-%s-%d", timestamp.Day(), timestamp.Month().String(), timestamp.Year())
}
