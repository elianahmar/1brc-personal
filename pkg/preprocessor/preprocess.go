package preprocessor

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

// TODO: this is where a majority of the optimizations will need to be made
// For context: we only process 413 cities in total. A majority of the time is gonna be from just
// reading the file
func ReadFile(path string, chanSize int) map[model.City]*model.Measurement {
	file := utils.PanicOnError(os.Open(path))
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	dataChan := make(chan string, chanSize)
	wg := &sync.WaitGroup{}
	wg.Add(2)
	// NOTE: this is where we read the line and push the line string to channel
	// Start one go routine which produces lines and pushes them to data
	// We start another go routine which is receiving from the data channel and updating a map
	// once the data chan is closed we will exit from collect data and push the map to the measurement chan
	// Simple producer consumer pattern
	go func() {
		defer wg.Done()
		lines := chanSize
		for fileScanner.Scan() {
			text := fileScanner.Text()
			dataChan <- text
			lines -= 1
			if lines <= 0 {
				break
			}
		}
		close(dataChan)
	}()

	measurementChan := make(chan map[model.City]*model.Measurement, 1)
	// I think for code clarify I make make this method return a func so I can just do go collectData()
	// That will clarify the code
	go collectData(dataChan, measurementChan, chanSize, wg)

	// measurements := collectData(data)
	wg.Wait()
	return <-measurementChan
}

func collectData(data chan string, measurementChan chan map[model.City]*model.Measurement, linesToProcess int, wg *sync.WaitGroup) {
	defer wg.Done()
	measurements := make(map[model.City]*model.Measurement, 500)
	linesProcessed := 0
	for text := range data {
		linesProcessed += 1
		city, temp := processLine(text)

		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{City: city}
		}
		measurements[city].Temps += temp
		measurements[city].Count += 1
		measurements[city].Max = math.Max(measurements[city].Max, temp)
		measurements[city].Min = math.Min(measurements[city].Min, temp)
	}
	utils.PanicOnCondition(linesProcessed != linesToProcess, "didn't process all lines")

	measurementChan <- measurements
	close(measurementChan)
}

func processLine(text string) (model.City, float64) {
	split := strings.Split(text, ";")
	dig := utils.PanicOnError(strconv.ParseFloat(split[1], 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.City(split[0]), temp
}
