package preprocessor

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

// Full Dataset -> 28 seconds
// Using this benchmark to see if it's faster to convert to int then convert to float
func Benchmark_IntFromFloat(b *testing.B) {
	parse := func(num []byte) (int, error) {
		numByte := make([]byte, 0, 8)
		if len(num) > 8 {
			b.Errorf("temperature should never exceed 8 bytes. Found = %d bytes", len(num))
		}
		for i := range num {
			nb := num[i]
			if nb == '.' {
				continue
			}
			numByte = append(numByte, num[i])
		}
		return strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
	}
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()
		mb := 1
		bufSize := mb * 1024 * 1024
		delim := []byte{';'}
		fileScanner := bufio.NewScanner(file)
		fileScanner.Buffer(make([]byte, bufSize), 1024)
		for fileScanner.Scan() {
			line := fileScanner.Bytes()
			_, num, _ := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
			_, err := parse(num)
			if err != nil {
				b.Errorf("failed to parse the num, %v", err)
			}
			// process the line itself
		}
	}
}

// Full Dataset -> 45 seconds
// Using this benchmark to see if it's faster to convert to int then convert to float
func Benchmark_IntFromFloatFullCompute(b *testing.B) {
	parse := func(num []byte) (int, error) {
		numByte := make([]byte, 0, 8)
		if len(num) > 8 {
			b.Errorf("temperature should never exceed 8 bytes. Found = %d bytes", len(num))
		}
		for i := range num {
			nb := num[i]
			if nb == '.' {
				continue
			}
			numByte = append(numByte, num[i])
		}
		return strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
	}
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()
		mb := 1
		bufSize := mb * 1024 * 1024
		delim := []byte{';'}
		measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
		fileScanner := bufio.NewScanner(file)
		fileScanner.Buffer(make([]byte, bufSize), 1024)
		for fileScanner.Scan() {
			line := fileScanner.Bytes()
			city, num, _ := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
			temp, err := parse(num)
			if err != nil {
				b.Errorf("failed to parse the num, %v", err)
			}
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
	}
}

// Full Dataset -> 48 seconds. Slower. Thinking that maybe we parse the full line instead
func Benchmark_IntFromFloatCut(b *testing.B) { //
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()
		mb := 1
		bufSize := mb * 1024 * 1024
		delim := []byte{';'}
		period := []byte{'.'}
		measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
		fileScanner := bufio.NewScanner(file)
		fileScanner.Buffer(make([]byte, bufSize), 1024)
		for fileScanner.Scan() {
			line := fileScanner.Bytes()
			city, num, _ := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
			d1, d2, _ := bytes.Cut(num, period)
			// temp, err := parse(num)
			// if err != nil {
			// 	b.Errorf("failed to parse the num, %v", err)
			// }
			d1 = append(d1, d2...)
			temp, _ := strconv.Atoi(unsafe.String(&d1[0], len(d1)))
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
	}
}
