package main

import (
	"fmt"
	"os"
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
	start := time.Now()
	runCalculations()
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
}

func runCalculations() {
	// PProf
	cpuProfile := utils.PanicOnError(os.Create("cpu.prof"))
	memProfile := utils.PanicOnError(os.Create("mem.prof"))
	defer func(cpuProfile *os.File, memProfile *os.File) {
		cpuProfile.Close()
		memProfile.Close()
	}(cpuProfile, memProfile)

	utils.PanicOnError(struct{}{}, pprof.StartCPUProfile(cpuProfile))
	defer pprof.StopCPUProfile()

	// Run the script
	measurements := pre.ReadFile("../1brc-go/measurements.txt", 1000)
	compute.ComputeAvg(measurements)
	validator.ValidateCorrectness(measurements)

	utils.PanicOnError(struct{}{}, pprof.Lookup("heap").WriteTo(memProfile, 0))
}
