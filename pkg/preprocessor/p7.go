package preprocessor

import (
	"bufio"
	"bytes"
	"os"
	"strconv"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P7 struct {
	Path     string
	ChanSize int
}

func NewP7(path string, chansize int) *P7 {
	return &P7{
		Path:     path,
		ChanSize: chansize,
	}
}

func (p7 *P7) Compute() map[string]*model.Measurement { // 53 seconds. Minimal difference using unsafe for temperature
	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p7.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Buffer(make([]byte, 2*1024*1024), 1024*1024)
	delim := []byte{';'}
	measurements := make(map[string]*model.Measurement, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		// process the line itself
		city, num, _ := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
		temp, _ := strconv.ParseFloat(utils.BytesToString(num), 64)
		measurement, exists := measurements[utils.BytesToString(city)] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			cityName := string(city)
			measurement = &model.Measurement{City: cityName}
			measurements[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	return measurements
}
