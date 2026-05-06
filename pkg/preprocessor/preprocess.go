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
		city, temp := processLine(text)

		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{City: city}
		}
		measurements[city].Temps += temp
		measurements[city].Count += 1
		measurements[city].Max = math.Max(measurements[city].Max, temp)
		measurements[city].Min = math.Min(measurements[city].Min, temp)
		// fmt.Printf("%v\n", measurements[city])
	}
	return measurements
}

func processLine(text string) (model.City, float64) {
	split := strings.Split(text, ";")
	dig := utils.PanicOnError(strconv.ParseFloat(split[1], 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.City(split[0]), temp
}
