package preprocessor

import (
	"bufio"
	"os"
	"strconv"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P10 struct {
	Path     string
	ChanSize int
}

func NewP10(path string) *P10 {
	return &P10{
		Path: path,
	}
}

func (p10 *P10) Compute() map[string]*model.MeasurementInt { // 108 seconds. New Record
	// Inlining this function to keep everything on the stack
	numByte := make([]byte, 0, 8)
	cityByte := make([]byte, 0, 32)
	delim, period := byte(';'), byte('.')
	parse := func(line []byte) (int, string) {
		numByte = numByte[:0]   // clear the array
		cityByte = cityByte[:0] // clear the array
		L := 0
		N := len(line)
		for {
			nb := line[L]
			if nb == delim {
				L += 1
				break
			}
			cityByte = append(cityByte, nb)
			L += 1
		}
		for L < N {
			nb := line[L]
			if nb == period {
				L += 1
				continue
			} else {
				numByte = append(numByte, nb)
				L += 1
			}
		}
		temp, _ := strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		city := unsafe.String(&cityByte[0], len(cityByte))
		return temp, city
	}
	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p10.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Buffer(make([]byte, 2*1024*1024), 1024*1024)
	measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		// process the line itself
		temp, city := parse(line)
		measurement, exists := measurements[city] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			cityName := string(city)
			measurement = &model.MeasurementInt{City: cityName}
			measurements[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		// PERF: Would min and max work on the strings themselves?
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	return measurements
}

// NOTE: Personal note about floating point representation in golang
// for float32, [1][8][23] => sign, exponent, fraction respectively
// for float64, [1][11][52] => sign, exponent, fraction respectively
