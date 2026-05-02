package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"
)

type measurement struct {
	city string
	temp float64
}

// TODO:
// - Read in the file line by line -> DONE
// - Cities appear multiple times
// - Implement the brute force solution first then track the time it takes

func main() {
	start := time.Now()
	fmt.Println("Hello World")
	parseFile()
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
}

func parseFile() {
	// First we need to read the file into an object
	readFile, err := os.Open("../measurements.txt")
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	lines := 1000
	data := make([]measurement, 0, 1000) // I can't even allocate an array of 1billion
	for fileScanner.Scan() {
		text := fileScanner.Text()
		measurement := processLine(text)
		fmt.Printf("%v\n", measurement)
		data = append(data, measurement)
		lines -= 1
		if lines <= 0 {
			break
		}
	}
	readFile.Close()
}

func processLine(text string) measurement {
	split := strings.Split(text, ";")
	dig, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		panic(err)
	}
	temp := truncateNaive(dig, 0.1) // No good. We don't need this much precision
	return measurement{
		city: split[0],
		temp: temp,
	}
}

func truncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}
