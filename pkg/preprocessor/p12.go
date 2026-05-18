package preprocessor

import (
	"bufio"
	"os"
	"strconv"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P12 struct {
	Path     string
	ChanSize int
}

func NewP12(path string) *P12 {
	return &P12{
		Path: path,
	}
}

func (p12 *P12) Compute() map[string]*model.MeasurementInt { // 44 seconds. I think I need to override some implementation

	// Inlining this function to keep everything on the stack... Is this actually the case?
	numByte := make([]byte, 0, 8)
	newline, delim, period := byte('\n'), byte(';'), byte('.')
	L, N, temp := 0, 0, 0

	// NOTE: Inlining the function doesn't improve speed. I think compiler is probably doing it for me
	parse := func(line []byte) (int, int) {
		numByte = numByte[:0] // clear the array
		L, N = 0, len(line)
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
		temp, _ = strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		return temp, delimIdx
	}

	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p12.Path))
	// defer file.Close() //NOTE: commenting this out saves a ~second
	reader := bufio.NewReaderSize(file, 2*1024*1024)
	measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
	for {
		line, err := reader.ReadSlice(newline) // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
		if err != nil {
			break
		}
		temp, delimIdx := parse(line)
		measurement, exists := measurements[unsafe.String(&line[0], delimIdx)] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			// NOTE: Was casting string to string which doesn't copy. That's why map data was wrong
			cityName := string(line[0:delimIdx])
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
