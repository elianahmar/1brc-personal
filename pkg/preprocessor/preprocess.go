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

type P1 struct {
	Path     string
	ChanSize int
}

func NewP1(path string, chansize int) *P1 {
	return &P1{
		Path:     path,
		ChanSize: chansize,
	}
}

// THOUGHTS: To read concurrently, I will need to read each byte individually. However, that will pose another problem
// If I chunk based on bytes, then there is a possibility of some lines being cut off. I would have to resolve those lines
// Let me think about this. I read the entire line by line and create an object for each line. What if read in parallel, rejoin the entire
// TODO: Come back to this. Let's first use a pool of strings.

// TODO: this is where a majority of the optimizations will need to be made
// For context: we only process 413 cities in total. A majority of the time is gonna be from just
// reading the file
func (p1 *P1) ReadFile(path string, chanSize int) map[model.City]*model.Measurement {
	dataChan := make(chan string, chanSize)
	wg := &sync.WaitGroup{}
	file := utils.PanicE(os.Open(path))
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	// NOTE: this is where we read the line and push the line string to channel
	// Start one go routine which produces lines and pushes them to data
	// We start another go routine which is receiving from the data channel and updating a map
	// once the data chan is closed we will exit from collect data and push the map to the measurement chan
	// Simple producer consumer pattern
	measurementChan := make(chan map[model.City]*model.Measurement, 1)

	wg.Add(2)
	// Producer consumer pattern. Consumers will stop receiving once the channel is closed
	go p1.pushLines(fileScanner, dataChan, chanSize, wg)
	go p1.collectData(dataChan, measurementChan, chanSize, wg)
	wg.Wait()

	return <-measurementChan
}

func (p1 *P1) collectData(data chan string, measurementChan chan map[model.City]*model.Measurement, linesToProcess int, wg *sync.WaitGroup) {
	defer wg.Done()
	measurements := make(map[model.City]*model.Measurement, 500)
	linesProcessed := 0
	for text := range data {
		linesProcessed += 1
		city, temp := p1.processLine(text)

		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{City: city}
		}
		measurements[city].Temps += temp
		measurements[city].Count += 1
		measurements[city].Max = math.Max(measurements[city].Max, temp)
		measurements[city].Min = math.Min(measurements[city].Min, temp)
	}
	utils.PanicIf(linesProcessed != linesToProcess, "didn't process all lines")

	measurementChan <- measurements
	close(measurementChan)
}

func (p1 *P1) pushLines(fileScanner *bufio.Scanner, dataChan chan string, chanSize int, wg *sync.WaitGroup) {
	defer wg.Done()
	p1.naiveLineScanner(fileScanner, dataChan, chanSize)
	close(dataChan)
}

func (p1 *P1) naiveLineScanner(fileScanner *bufio.Scanner, dataChan chan string, chanSize int) {
	lines := chanSize
	for fileScanner.Scan() {
		text := fileScanner.Text()
		dataChan <- text
		lines -= 1
		if lines <= 0 {
			break
		}
	}
}

func (p1 *P1) processLine(text string) (model.City, float64) {
	split := strings.Split(text, ";")
	dig := utils.PanicE(strconv.ParseFloat(split[1], 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.City(split[0]), temp
}
