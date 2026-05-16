package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strconv"
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

	var chansize *int
	if len(os.Args) > 3 {
		num := utils.PanicE(strconv.Atoi(os.Args[3]))
		chansize = &num
	} else {
		defChanSize := 1000000000
		chansize = &defChanSize
	}
	runImplementation(os.Args[1], os.Args[2], chansize)
	utils.PanicE(struct{}{}, pprof.WriteHeapProfile(memProfile))
}

func runImplementation(impl, path string, chansize *int) {
	switch impl {
	// case "p1":
	// 	return pre.NewP1(path, *chansize)
	case "p2":
		panic("not implemented")
	// case "p3":
	// 	return pre.NewP3(path)
	case "p4":
		fmt.Println("Read the file and processed the lines")
		p4 := pre.NewP4(path, *chansize)
		measurements := p4.Compute()
		compute.ComputeAvg(measurements)
		validator.ValidateCorrectness(measurements)
	case "p5":
		fmt.Println("Read the file and processed the lines")
		p5 := pre.NewP5(path, *chansize)
		measurements := p5.Compute()
		compute.ComputeAvg(measurements)
		validator.ValidateCorrectness(measurements)

	case "p6":
		fmt.Println("Read the file and processed the lines")
		p6 := pre.NewP6(path, *chansize)
		measurements := p6.Compute()
		compute.ComputeAvg(measurements)
		validator.ValidateCorrectness(measurements)

	case "p7":
		fmt.Println("Read the file and processed the lines")
		p7 := pre.NewP7(path)
		measurements := p7.Compute()
		compute.ComputeAvg(measurements)
		validator.ValidateCorrectness(measurements)
	case "p8":
		fmt.Println("Read the file and processed the lines")
		p8 := pre.NewP8(path)
		measurements := p8.Compute()
		predictions := compute.ComputeAvgStrConv(measurements)
		fmt.Println("Computed the averages. Time to validate")
		validator.ValidateCorrectnessInt(predictions)
	case "p9":
		fmt.Println("Read the file and processed the lines")
		p9 := pre.NewP9(path)
		measurements := p9.Compute()
		predictions := compute.ComputeAvgStrConv(measurements)
		fmt.Println("Computed the averages. Time to validate")
		validator.ValidateCorrectnessInt(predictions)
	case "p10":
		fmt.Println("Read the file and processed the lines")
		p10 := pre.NewP10(path)
		measurements := p10.Compute()
		predictions := compute.ComputeAvgStrConv(measurements)
		fmt.Println("Computed the averages. Time to validate")
		validator.ValidateCorrectnessInt(predictions)

	}
}
