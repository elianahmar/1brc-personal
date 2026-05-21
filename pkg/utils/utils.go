package utils

import (
	"fmt"
	"math"
	"time"
	"unsafe"
)

func TruncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}

func PanicE[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

func PanicIf(cond bool, msg string) {
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

// NOTE: Source : https://dev.to/devflex-pro/how-to-use-unsafe-in-go-without-killing-your-service-699
// With unsafe, I'm taking creating a string header that points to the byte array. This should only be used
// in cases where the underlying byte array isn't being mutated. The upside to this is that there is zero-allocation
// Because casting b to string creates a copy of the byte array O(n). With unsafe pointers this becomes O(1)
func BytesToString(b []byte) string {
	return unsafe.String(&b[0], len(b))
}

func PrintMap[T comparable, E any](m map[T]E) {
	println("\n++++++ PRINTING MAP CONTENTS ++++++\n")
	for k, v := range m {
		println("key = ", k, "value = ", v)
	}
	println("\n++++++ PRINTING MAP CONTENTS ++++++\n")
}
