package cmd2

import (
	"fmt"
	"time"

	"github.com/throwea/1brc-go/pkg/validator"
)

func main() {
	start := time.Now()
	measurements := readFile()
	measurements := calculate(measurements)
	fmt.Printf("Time taken: %2f", time.Since(start).Seconds())
	validator.ValidateCorrectness(measurements)
}
