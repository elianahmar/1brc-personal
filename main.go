package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/throwea/1brc-go/pkg/compute"
	"github.com/throwea/1brc-go/pkg/files"
	pre "github.com/throwea/1brc-go/pkg/preprocessor"
	"github.com/throwea/1brc-go/pkg/utils"
	"github.com/throwea/1brc-go/pkg/validator"
)

// TODO:
// - Optional. Add some command line args for channel size
// - append date to cpu.prof, mem.prof so I have multiple dumps to see progress -> DONE
// - Move the dumps to directories that are titled with day month year, then append the seconds so we can see clearly

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
	files.CreateDir(dayMonthYear)
	cpuFile := fmt.Sprintf("./documentation/%s/%s-cpu.prof", dayMonthYear, dayMonthYear)
	memFile := fmt.Sprintf("./documentation/%s/%s-mem.prof", dayMonthYear, dayMonthYear)

	cpuProfile := utils.PanicE(os.Create(cpuFile))
	memProfile := utils.PanicE(os.Create(memFile))
	defer func(cpuProfile *os.File, memProfile *os.File) {
		cpuProfile.Close()
		memProfile.Close()
	}(cpuProfile, memProfile)

	utils.PanicE(struct{}{}, pprof.StartCPUProfile(cpuProfile))
	defer pprof.StopCPUProfile()

	p3 := pre.NewP3("../1brc-go/small_measurements.txt")
	measurements := p3.ReadFileConcurrent()
	fmt.Println("Read the file and processed the lines")
	compute.ComputeAvg(measurements)
	fmt.Println("Computed the averages. Time to validate")
	validator.ValidateCorrectness(measurements)
	fmt.Println("Finished validating the answers")
	utils.PanicE(struct{}{}, pprof.WriteHeapProfile(memProfile))
}

func selectImplementation(impl string) model.Compute {
	switch impl {
	case "p1":
	case "p2":
	case "p3":
	case "p4":
	case "p5":

	}
}
