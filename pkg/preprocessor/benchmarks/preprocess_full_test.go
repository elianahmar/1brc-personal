package preprocessor

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"testing"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/preprocessor"
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

func Benchmark_SimpleParserFull(b *testing.B) { // 108 seconds. New Record
	// Inlining this function to keep everything on the stack
	for b.Loop() {
		fmt.Println("begin")
		numByte := make([]byte, 0, 8)
		cityByte := make([]byte, 0, 32)
		parse := func(line []byte) (int, string) {
			numByte = numByte[:0]   // clear the array
			cityByte = cityByte[:0] // clear the array
			L := 0
			for {
				nb := line[L]
				if nb == ';' {
					L += 1
					break
				}
				numByte = append(numByte, nb)
			}
			for i := L; i < len(line); i++ {
				nb := line[i]
				if nb == '.' {
					continue
				} else if nb == '\n' {
					break
				} else {
					cityByte = append(cityByte, nb)
				}
			}
			fmt.Println("reached the end")
			temp, _ := strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
			return temp, unsafe.String(&cityByte[0], len(cityByte))
		}
		// Brute force this. Read line by line and update a table
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
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
			// PERF: Would min and max work on the strings themselves?
			measurement.Max = max(measurement.Max, temp)
			measurement.Min = min(measurement.Min, temp)
		}
	}
}

func Benchmark_SimpleParserIndexByte(b *testing.B) {
	// Inlining this function to keep everything on the stack
	numByte := make([]byte, 0, 8)
	delim, period := byte(';'), byte('.')
	L, N, temp := 0, 0, 0

	// NOTE: Inlining the function doesn't improve speed. I think compiler is probably doing it for me
	parse := func(line []byte) (int, int) {
		numByte = numByte[:0] // clear the array
		N = len(line)
		delimIdx := bytes.IndexByte(line, delim)
		L = delimIdx + 1
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
	for b.Loop() {

		fmt.Println("begin")
		file := utils.PanicE(os.Open("../../../../1brc-go/measurements.txt"))
		defer file.Close()
		fileScanner := bufio.NewScanner(file)
		fileScanner.Buffer(make([]byte, 2*1024*1024), 1024*1024)
		measurements := make(map[string]*model.MeasurementInt, 512) // 512 bc it's power of 2
		for fileScanner.Scan() {

			line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
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
	}
}

func Benchmark_P12(b *testing.B) { // 44.30 seconds
	p12 := preprocessor.NewP12("../../../../1brc-go/measurements.txt")
	for b.Loop() {
		p12.Compute()
	}
}

func Benchmark_P11(b *testing.B) { // 39.82 seconds
	p11 := preprocessor.NewP11("../../../../1brc-go/measurements.txt")
	for b.Loop() {
		p11.Compute()
	}
}
