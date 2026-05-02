package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
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
// - I'm thinking here that I should process and push to a channel in two separate go routines
//		since I can't allocate an array this big

func main() {
	start := time.Now()
	fmt.Println("Hello World")
	parseFile()
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
}

func parseFile() {
	// First we need to read the file into an object
	readFile, err := os.Open("../1brc-go/measurements.txt")
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	data := make(chan string, 1000)
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		lines := 1000
		for fileScanner.Scan() {
			text := fileScanner.Text()
			data <- text
			// fmt.Printf("\n%d", lines)
			lines -= 1
			if lines <= 0 {
				break
			}
		}
		close(data)
	}()

	go func(data chan string) {
		defer wg.Done()
		for text := range data {
			measurement := processLine(text)
			fmt.Printf("%v\n", measurement)
		}
	}(data)
	wg.Wait()
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
