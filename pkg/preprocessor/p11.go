package preprocessor

import (
	"bufio"
	"os"
	"strconv"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P11 struct {
	Path     string
	ChanSize int
}

func NewP11(path string) *P11 {
	return &P11{
		Path: path,
	}
}

func (p11 *P11) Compute() map[string]*model.MeasurementInt { // 38 seconds.

	// Inlining this function to keep everything on the stack
	numByte := make([]byte, 0, 8)
	delim, period := byte(';'), byte('.')

	parse := func(line []byte) (int, int) {
		numByte = numByte[:0] // clear the array
		L, N := 0, len(line)
		for line[L] != delim {
			L += 1
		}
		delimIdx := L
		L += 1
		for L < N {
			nb := line[L]
			if nb != period {
				numByte = append(numByte, nb)
			}
			L += 1
		}
		// NOTE: Just had this idea. Might be able to remove numByte and CityByte array
		// entirely and just do unsafe string on the length and find the index of the ';' char
		// In future attempts, might just be able to override scanner implementation. I think they expose the interfaces
		temp, _ := strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		return temp, delimIdx
	}

	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p11.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	fileScanner.Buffer(make([]byte, 2*1024*1024), 1024*1024)
	measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		temp, delimIdx := parse(line)
		city := unsafe.String(&line[0], delimIdx)
		measurement, exists := measurements[city] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			// NOTE: Was casting string to string which doesn't copy. That's why map data was wrong
			cityName := string(line[0:delimIdx])
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
