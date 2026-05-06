package cmd2

import (
	"fmt"
	"time"
)

func main() {
	start := time.Now()
	measurements := readFile()
	measurements := calculate(measurements)
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
	validateCorrectness(measurements)
}
