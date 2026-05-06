package main

import (
	"fmt"
	"time"

	"github.com/throwea/1brc-go/pkg/compute"
	pre "github.com/throwea/1brc-go/pkg/preprocessor"
	"github.com/throwea/1brc-go/pkg/validator"
)

// TODO:
// - No matter how slow, get a working solution to have a baseline

func main() {
	start := time.Now()
	runCalculations()
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
}

func runCalculations() {
	measurements := pre.ReadFile("../../1brc-go/measurements.txt", 100000000)

	compute.ComputeAvg(measurements)

	validator.ValidateCorrectness(measurements)
}
