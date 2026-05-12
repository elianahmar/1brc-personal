package preprocessor

import (
	"bufio"
	"bytes"
	"math"
	"os"
	"strconv"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P5 struct {
	Path     string
	ChanSize int
}

func NewP5(path string, chansize int) *P5 {
	return &P5{
		Path:     path,
		ChanSize: chansize,
	}
}

// THOUGHTS: To read concurrently, I will need to read each byte individually. However, that will pose another problem
// If I chunk based on bytes, then there is a possibility of some lines being cut off. I would have to resolve those lines
// Let me think about this. I read the entire line by line and create an object for each line. What if read in parallel, rejoin the entire

func (p5 *P5) Compute() map[string]*model.Measurement { // 56 seconds. Fastest yet. All single threaded????
	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p5.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	delim := []byte{';'}
	measurements := make(map[string]*model.Measurement, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		// process the line itself
		city, num, found := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
		cityLookup := utils.BytesToString(city)
		utils.PanicIf(!found, "bytes not found?")
		temp := utils.PanicE(strconv.ParseFloat(string(num), 64))
		measurement, exists := measurements[cityLookup] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			cityName := string(city)
			measurement = &model.Measurement{City: cityName}
			measurements[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = math.Max(measurement.Max, temp)
		measurement.Min = math.Min(measurement.Min, temp)
	}
	return measurements
}
