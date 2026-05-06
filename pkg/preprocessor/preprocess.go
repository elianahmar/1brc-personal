package preprocessor

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

// TODO: add command line args for testing purposes
func ReadFile(path string, chanSize int) map[model.City]*model.Measurement {
	file := utils.PanicOnError(os.Open(path))
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	data := make(chan string, chanSize)
	wg := sync.WaitGroup{}
	wg.Add(1)
	// NOTE: this is where we read the line and push the line string to channel
	go func() {
		defer wg.Done()
		lines := chanSize
		for fileScanner.Scan() {
			text := fileScanner.Text()
			data <- text
			lines -= 1
			if lines <= 0 {
				break
			}
		}
		close(data)
	}()
	measurements := collectData(data)
	wg.Wait()
	return measurements
}

func collectData(data chan string) map[model.City]*model.Measurement {
	measurements := make(map[model.City]*model.Measurement)
	for text := range data {
		measurement := processLine(text)
		split := strings.Split(text, ";")
		city := model.City(split[0])
		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{}
		}
		newTemp := measurement.Temps
		measurements[city].Temps += newTemp
		measurements[city].Count += 1
		measurements[city].Max = math.Max(measurements[city].Max, newTemp)
		measurements[city].Min = math.Min(measurements[city].Min, newTemp)
		fmt.Printf("%v\n", measurement)
	}
	return measurements
}

func processLine(text string) model.Measurement {
	split := strings.Split(text, ";")
	dig, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		panic(err)
	}
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.Measurement{
		City:  split[0],
		Temps: temp,
	}
}
