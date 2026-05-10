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
// - Optional. Add some command line args for channel size
// - append date to cpu.prof, mem.prof so I have multiple dumps to see progress

func main() {
	// Start CPU and Memory Profiling
	// runtime.SetCPUProfileRate(1)
	runtime.SetBlockProfileRate(1)

	start := time.Now()
	runCalculations()
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
}

func runCalculations() {
	dayMonthYear := utils.DayMonthYear()
	cpuFile := fmt.Sprintf("%s-cpu.prof", dayMonthYear)
	memFile := fmt.Sprintf("%s-mem.prof", dayMonthYear)

	cpuProfile := utils.PanicOnError(os.Create(cpuFile))
	memProfile := utils.PanicOnError(os.Create(memFile))
	defer func(cpuProfile *os.File, memProfile *os.File) {
		cpuProfile.Close()
		memProfile.Close()
	}(cpuProfile, memProfile)

	utils.PanicOnError(struct{}{}, pprof.StartCPUProfile(cpuProfile))
	defer pprof.StopCPUProfile()

	measurements := pre.ReadFileConcurrent2("../1brc-go/measurements.txt")
	fmt.Println("Read the file and processed the lines")
	compute.ComputeAvg(measurements)
	fmt.Println("Computed the averages. Time to validate")
	validator.ValidateCorrectness(measurements)
	fmt.Println("Finished validating the answers")
	utils.PanicOnError(struct{}{}, pprof.WriteHeapProfile(memProfile))
}
