package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/throwea/1brc-go/pkg/compute"
	pre "github.com/throwea/1brc-go/pkg/preprocessor"
	"github.com/throwea/1brc-go/pkg/utils"
	"github.com/throwea/1brc-go/pkg/validator"
)

// TODO:
// - No matter how slow, get a working solution to have a baseline
// - Optional. Add some command line args for channel size

func main() {
	// Start CPU and Memory Profiling
	// runtime.SetCPUProfileRate(1)
	runtime.SetBlockProfileRate(1)
	cpuProfile := utils.PanicOnError(os.Create("cpu.prof"))
	memProfile := utils.PanicOnError(os.Create("mem.prof"))
	utils.PanicOnError(struct{}{}, pprof.StartCPUProfile(cpuProfile))
	defer func(cpuProfile *os.File, memProfile *os.File) {
		cpuProfile.Close()
		memProfile.Close()
		pprof.StopCPUProfile()
	}(cpuProfile, memProfile)

	start := time.Now()

	runCalculations()

	utils.PanicOnError(struct{}{}, pprof.WriteHeapProfile(memProfile))
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
}

func runCalculations() {
	measurements := pre.ReadFileConcurrent2("../1brc-go/measurements.txt")
	compute.ComputeAvg(measurements)
	validator.ValidateCorrectness(measurements)
}
