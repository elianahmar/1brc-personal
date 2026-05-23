package preprocessor

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func Test_SimpleParser(t *testing.T) {
	period, delim := byte('.'), byte(';')
	numByte := make([]byte, 0, 8)
	cityByte := make([]byte, 0, 32)
	parse := func(line []byte) (int, string) {
		numByte = numByte[:0]   // clear the array
		cityByte = cityByte[:0] // clear the array
		L := 0
		for {
			nb := line[L]
			if nb == delim {
				L += 1
				break
			}
			cityByte = append(cityByte, nb)
			L += 1
		}
		fmt.Printf("finished parsing the city, L = %d\n", L)

		for L < len(line) {
			nb := line[L]
			if nb != period {
				numByte = append(numByte, nb)
			}
			L += 1
		}
		fmt.Println("string numbyte: ", string(numByte))
		fmt.Println("string citybyte: ", string(cityByte))
		temp, _ := strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		return temp, unsafe.String(&cityByte[0], len(cityByte))
	}
	line := []byte("Baltimore;12.0")
	fmt.Printf("line length = %d\n", len(line))
	temp, city := parse(line)
	fmt.Println("temp = ", temp)
	fmt.Println("city = ", city)
	fmt.Printf("Full Line = %s;%d\n", city, temp)
	if !strings.EqualFold(city, "Baltimore") {
		t.Errorf("city is not correct; expected: Baltimore, actual: %s", city)
	}
	if temp != 120 {
		t.Errorf("temp is not correct; expected: 120, actual: %d", temp)
	}
}

func Test_ParseNum(t *testing.T) {
	delim := byte(';')
	numByte := make([]byte, 0, 8)
	parse := func(num []byte) int {
		numByte = numByte[:0] // clear the array
		L := 0
		for {
			nb := num[L]
			if nb == delim {
				L += 1
				break
			}
			numByte = append(numByte, nb)
			L += 1
		}
		fmt.Println("finished parsing the city")

		temp, _ := strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		return temp
	}
	temp := parse([]byte("Baltimore;12.0"))
	fmt.Println("temp + ", temp)
	if temp == 120 {
		t.Errorf("temp is not correct; expected: 120, actual: %d", temp)
	}
}

// func Test_FindDelimIdx(t *testing.T) {
// 	delim := byte(';')
// 	parse := func(line []byte) int {
// 		L := 0
// 		for {
// 			nb := line[L]
// 			if nb == delim {
// 				break
// 			}
// 			L += 1
// 		}
// 		return L
// 	}
// 	line := []byte("Baltimore;12.0\n")
// 	N := len(line)
// 	delimIdx := parse(line)
// 	city := unsafe.String(&line[0], delimIdx)
// 	fmt.Printf("city = %s\n", city)
// 	temp, _ := strconv.Atoi(unsafe.String(&line[delimIdx+1], N))
//
// 	fmt.Println("temp + ", temp)
// 	if !strings.EqualFold(city, "Baltimore") {
// 		t.Errorf("city is not correct; expected: Baltimore, actual: %s", city)
// 	}
// 	if temp != 120 {
// 		t.Errorf("temp is not correct; expected: 120, actual: %d", temp)
// 	}
// }

func Test_ParseNumNoConv(t *testing.T) {
	testCases := []struct {
		input    []byte
		desc     string
		expected int
	}{
		{
			desc:     "expecting 1.0",
			expected: 10,
			input:    []byte("1.0"),
		},
		{
			desc:     "expecting -13.5",
			expected: -135,
			input:    []byte("-13.5"),
		},
		{
			desc:     "expecting 10.5",
			expected: 105,
			input:    []byte("10.5"),
		},
		{
			desc:     "expecting -1.5",
			expected: -15,
			input:    []byte("-1.5"),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value := numConvertor(tC.input)
			println("\n", value)
			if value != tC.expected {
				t.Errorf("expected = %d, actual = %d", tC.expected, value)
			}
		})
	}
}

// -12.0
// - skips
// 1 -> 1 * 10 + 1 == 11
// 2 -> 11 * 10 + 2 == 11
//
// temp = 0
// 1 -> 0 * 10 + 1 = 1
// 2 -> 1 * 10 + 2 = 12
// 0 -> 12 * 10 + 0 = 120
// NOTE:
// We are converting the ascii digit byte into an integer
// However, we can't just cast byte of number to int.
// Because '0' -> '9' have int values 48 - 57
// So I have to take int(char - '0') which internally
// Gives me the correct numeric conversion
// TODO: Benchmark this
func numConvertor(numByte []byte) int {
	res := 0
	zero, nine, negative := byte('0'), byte('9'), byte('-')
	for _, char := range numByte {
		isDig := char >= zero && char <= nine // Fastest way I could find to tell if byte is digit. Use ascii comparison. No rune conversion
		if !isDig {
			continue
		}
		res *= 10
		res += int(char - zero)
	}
	isNeg := numByte[0] == negative
	if isNeg {
		return -1 * res
	}
	return res
}

func Benchmark_BoundsWhile(b *testing.B) { // 263.9 ns/op
	for b.Loop() {
		L, N := 0, 1000
		for L < N {
			L++
		}
	}
}

func Benchmark_ForLoop(b *testing.B) { // 266.3
	for b.Loop() {
		N := 1000
		for L := range N {
			_ = L
		}
	}
}

func Test_ParseMiddleLine(t *testing.T) {
	line := []byte("Baltimore;-13.0\nNew York City;-1.0\nChicago;2.0\n")

	temp, city, ptr, delimIdx, newLineFound := ParseLine(16, line, len(line))
	println(city, 16, delimIdx)
	if temp != -10 {
		t.Errorf("wrong temparature, got %d, expected %d\n", temp, -130)
	} else if city != "New York City" {
		t.Errorf("wrong city, got %s", city)
	} else if ptr != 35 {
		t.Errorf("ptr location is wrong, got %d, expected = %d", ptr, 33)
	} else if delimIdx != 29 {
		t.Errorf("delimIdx location is wrong")
	} else if !newLineFound {
		t.Errorf("newline not found")
	}
}

func Test_ParseSingleLine(t *testing.T) {
	line := []byte("Baltimore;-13.0\n")

	temp, city, ptr, delimIdx, newLineFound := ParseLine(0, line, len(line))
	if temp != -130 {
		t.Errorf("wrong temparature, got %d, expected %d", temp, -130)
	} else if city != "Baltimore" {
		t.Errorf("wrong city, got %s", city)
	} else if ptr != len(line) {
		t.Errorf("ptr location is wrong")
	} else if delimIdx != 9 {
		t.Errorf("delimIdx location is wrong")
	} else if !newLineFound {
		t.Errorf("newline not found")
	}
}

// TODO: still haven't asserted every range ends with '\n'
func Test_ParseSingleRange(t *testing.T) {
	buff := []byte("Baltimore;-13.0\nNew York City;-1.0\nChicago;2.0\n")
	ranges := model.Range{
		Start: 0,
		End:   46,
	}

	m := processRange(ranges, buff)
	// First make assertions on the map itself.
	utils.PrintMap(m)
	nYTemp, nyok := m["New York City"]
	chiTemp, cook := m["Chicago"]
	balTemp, bok := m["Baltimore"]
	if !nyok {
		t.Errorf("new york not found")
	} else if !cook {
		t.Errorf("chicago not found")
	} else if !bok {
		t.Errorf("baltimore not found")
	} else if nYTemp.Temps != -10 {
		t.Errorf("ny temp not correct")
	} else if chiTemp.Temps != 20 {
		t.Errorf("ny temp not correct")
	} else if balTemp.Temps != -130 {
		t.Errorf("ny temp not correct")
	}
}

// func Test_ParseMultipleRanges(t *testing.T) {
// 	buff := []byte("Baltimore;-13.0\nNew York City;-1.0\nChicago;2.0\n")
// 	ranges := []model.Range{
// 		{
// 			Start: 0,
// 			End:   15,
// 		},
// 		{
// 			Start: 16,
// 			End:   34,
// 		},
// 		{
// 			Start: 35,
// 			End:   46,
// 		},
// 	}
//
// 	temp, city, ptr, delimIdx, newLineFound := processRange(buff)
// 	//First make assertions on the map itself.
// 	// newYork, ok :=
// }

func ParseLine(start int, buff []byte, N int) (int, string, int, int, bool) {
	var (
		delim                = byte(';')
		newline              = byte('\n')
		zero, nine, negative = byte('0'), byte('9'), byte('-')
		temp, delimIdx       = 0, 0
		newLineFound         = false
	)

	ptr := start
	for ptr < N {
		if buff[ptr] == delim {
			delimIdx = ptr
			break
		}
		ptr++
	}

	ptr++ // move past the ';'
	temp = 0
	isNeg := buff[ptr] == negative
	for ptr < N {
		nb := buff[ptr]
		if nb == newline {
			newLineFound = true
			break
		}
		if zero <= nb && nb <= nine {
			temp = (temp * 10) + int(nb-zero)
		}
		ptr++
	}
	ptr++                                               // So that we move past the newline break
	city := unsafe.String(&buff[start], delimIdx-start) // BUG: This is printing the full line
	// utils.PanicIf(delimIdx-start <= 0, "indexes are off")
	// utils.PanicIf(strings.ContainsRune(city, rune(';')), city+" contains delimeter")
	if isNeg {
		temp *= -1
	}
	return temp, city, ptr, delimIdx, newLineFound
}

func processRange(r model.Range, buff []byte) map[string]*model.MeasurementInt {
	N, start := 0, 0

	localMeasurement := make(map[string]*model.MeasurementInt, 512)
	ptr := 0
	N = len(buff)
	for ptr < N {
		start = ptr
		temp, city, nlIdx, dlIdx, nlFound := ParseLine(ptr, buff, N)
		if !nlFound {
			break
		}
		ptr = nlIdx

		measurement, exists := localMeasurement[city] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			// NOTE: Was casting string to string which doesn't copy. That's why map data was wrong
			cityName := string(buff[start:dlIdx])
			measurement = &model.MeasurementInt{City: cityName}
			localMeasurement[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	return localMeasurement
}
