package preprocessor

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"unsafe"
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
			if nb == period {
				L += 1
				continue
			} else {
				numByte = append(numByte, nb)
				L += 1
			}
		}
		fmt.Println("string numbyte: ", string(numByte))
		fmt.Println("string citybyte: ", string(cityByte))
		temp, _ := strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		return temp, unsafe.String(&cityByte[0], len(cityByte))
	}
	line := []byte("Baltimore;12.0")
	fmt.Printf("line length = %d\n", len(line))
	temp, city := parse(line)
	fmt.Println("temp + ", temp)
	fmt.Println("city + ", city)
	fmt.Printf("%s;%d\n", city, temp)
	if strings.EqualFold(city, "Baltimore") {
		t.Errorf("city is not correct; expected: Baltimore, actual: %s", city)
	}
	if temp == 120 {
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
