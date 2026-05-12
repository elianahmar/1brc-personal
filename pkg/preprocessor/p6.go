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

type P6 struct {
	Path     string
	ChanSize int
}

func NewP6(path string, chansize int) *P6 {
	return &P6{
		Path:     path,
		ChanSize: chansize,
	}
}

func (p6 *P6) Compute() map[string]*model.Measurement { // 55 seconds. Minimal difference using unsafe for temperature
	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p6.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Buffer(make([]byte, 2*1024*1024), 1024*1024)
	delim := []byte{';'}
	measurements := make(map[string]*model.Measurement, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		// process the line itself
		city, num, found := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
		cityLookup := utils.BytesToString(city)
		numUnsafe := utils.BytesToString(num)
		utils.PanicIf(!found, "bytes not found?")
		temp := utils.PanicE(strconv.ParseFloat(numUnsafe, 64))
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
