package cmd2

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/throwea/1brc-go/pkg/model"
)

type city string

func main() {
	start := time.Now()
	measurements := readFile()
	measurements := calculate(measurements)
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
	validateCorrectness(measurements)
}

func readFile() map[city]model.Measurement {
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
			lines -= 1
			if lines <= 0 {
				break
			}
		}
		close(data)
	}()
	measurements := collectData()

	wg.Wait()
	readFile.Close()
	return measurements
}
