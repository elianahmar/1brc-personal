package preprocessor

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P8 struct {
	Path     string
	ChanSize int
}

func NewP8(path string) *P8 {
	return &P8{
		Path: path,
	}
}

func (p8 *P8) Compute() map[string]*model.MeasurementInt { // 108 seconds. Twice as slow now
	// Inlining this function to keep everything on the stack
	parse := func(num []byte) (int, error) {
		numByte := make([]byte, 0, 8) // If this ends up being faster, think about buffering this or clearing after use?
		for i := range num {
			nb := num[i]
			if nb == '.' {
				continue
			}
			numByte = append(numByte, nb)
		}
		// Remove this after validating correctness
		utils.PanicIf(len(numByte) > 8, fmt.Sprintf("numByte array should never exceed 8 bytes. Length = %d", len(numByte)))
		return strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
	}
	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p8.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Buffer(make([]byte, 2*1024*1024), 1024*1024)
	delim := []byte{';'}
	measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		// process the line itself
		city, num, _ := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
		temp, _ := parse(num)
		measurement, exists := measurements[unsafe.String(&city[0], len(city))] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			cityName := string(city)
			measurement = &model.MeasurementInt{City: cityName}
			measurements[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	return measurements
}

// NOTE: Personal note about floating point representation in golang
// for float32, [1][8][23] => sign, exponent, fraction respectively
// for float64, [1][11][52] => sign, exponent, fraction respectively
