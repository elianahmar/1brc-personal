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

func (p1 *P1) Compute() map[string]*model.Measurement { // 509 seconds (over 8 minutes)
	dataChan := make(chan string, p1.ChanSize)
	wg := &sync.WaitGroup{}
	file := utils.PanicE(os.Open(p1.Path))
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)

	// NOTE: this is where we read the line and push the line string to channel
	// Start one go routine which produces lines and pushes them to data
	// We start another go routine which is receiving from the data channel and updating a map
	// once the data chan is closed we will exit from collect data and push the map to the measurement chan
	// Simple producer consumer pattern
	measurementChan := make(chan map[string]*model.Measurement, 1)

	wg.Add(2)
	// Producer consumer pattern. Consumers will stop receiving once the channel is closed
	go p1.pushLines(fileScanner, dataChan, p1.ChanSize, wg)
	go p1.collectData(dataChan, measurementChan, p1.ChanSize, wg)
	wg.Wait()

	return <-measurementChan
}

func (p1 *P1) collectData(data chan string, measurementChan chan map[string]*model.Measurement, linesToProcess int, wg *sync.WaitGroup) {
	defer wg.Done()
	measurements := make(map[string]*model.Measurement, 500)
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
	utils.PanicIf(linesProcessed != linesToProcess, fmt.Sprintf("didn't process all lines %d/%d", linesProcessed, linesToProcess))

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

func (p1 *P1) processLine(text string) (string, float64) {
	split := strings.Split(text, ";")
	dig := utils.PanicE(strconv.ParseFloat(split[1], 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return string(split[0]), temp
}
