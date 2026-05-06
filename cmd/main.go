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

	"github.com/throwea/1brc-go/pkg/model"
)

// TODO:
// - Implement the brute force solution first then track the time it takes

func main() {
	start := time.Now()
	fmt.Println("Hello World")
	measurements := getMeasurements()
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
	validateCorrectness(measurements)
}

func getMeasurements() map[city]model.Measurement {
	// First we need to read the file into an object
	readFile, err := os.Open("../1brc-go/measurements.txt")
	if err != nil {
		panic(err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	data := make(chan string, 100000000)
	wg := sync.WaitGroup{}
	wg.Add(1)
	// NOTE: this is where we read the line and push the line string to channel
	go func() {
		defer wg.Done()
		lines := 1000000
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

	// go func(data chan string) {
	// defer wg.Done()
	// NOTE: consume from the channel.
	measurements := make(map[model.City]model.Measurement)
	for text := range data {
		measurement := processLine(text)
		split := strings.Split(text, ";")
		city := model.City(split[0])
		if _, exists := measurements[city]; !exists {
			measurements[city] = model.Measurement{}
		}
		measurements[city].Temps += measurement.Temps
		measurements[city].Count += 1
		fmt.Printf("%v\n", measurement)
	}
	// }(data)
	wg.Wait()
	readFile.Close()
	return measurements
}

func processLine(text string) model.Measurement {
	split := strings.Split(text, ";")
	dig, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		panic(err)
	}
	temp := truncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.Measurement{
		city: split[0],
		temp: temp,
	}
}

func truncateNaive(f float64, unit float64) float64 {
	return math.Trunc(f/unit) * unit
}
