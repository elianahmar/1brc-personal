package preprocessor

import (
	"bufio"
	"io"
	"os"
	"testing"

	"github.com/throwea/1brc-go/pkg/utils"
)

// TODO: test that the reconcile lines method actually works correctly
// - 1. Test p1 file reading
// - 2. Test p1 file reading
// - 3. Test p1 file reading
// - 4. Test p1 file reading
// - 5. Test p1 file reading

func BenchmarkFileReadP1(b *testing.B) { // 1.280s
	path := "../../../1brc-go/small_measurements.txt" // 10*6 lines
	file := utils.PanicE(os.Open(path))
	lines := 1000000
	fakeChan := make(chan string, lines)
	for b.Loop() {
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			text := fileScanner.Text()
			fakeChan <- text
			lines -= 1
			if lines <= 0 {
				break
			}
		}
	}
}

func BenchmarkFileReadP1_BytesVersion(b *testing.B) { // 1.424s
	path := "../../../1brc-go/small_measurements.txt" // 10*6 lines
	file := utils.PanicE(os.Open(path))
	lines := 1000000
	fakeChan := make(chan []byte, lines)
	for b.Loop() {
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			text := fileScanner.Bytes()
			fakeChan <- text
			lines -= 1
			if lines <= 0 {
				break
			}
		}
	}
}

func BenchmarkFileReadP1_BytesVersion_NoChan(b *testing.B) { // 1.464s
	path := "../../../1brc-go/small_measurements.txt" // 10*6 lines
	file := utils.PanicE(os.Open(path))
	lines := 1000000
	for b.Loop() {
		fileScanner := bufio.NewScanner(file)
		fileScanner.Split(bufio.ScanLines)
		for fileScanner.Scan() {
			fileScanner.Bytes()
			lines -= 1
			if lines <= 0 {
				break
			}
		}
	}
}

func BenchmarkCyclicBuffer(b *testing.B) { // 1.465s (10*6 lines)
	for b.Loop() {
		path := "../../../1brc-go/small_measurements.txt" // 10*6 lines
		file := utils.PanicE(os.Open(path))
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			buf := make([]byte, 4*1024) // the chunk size
			// NOTE: n is the number of bytes read into the buffer
			// So the zero check is if we haven't read anything into the buffer
			// Reader must internally keep track of it's location as it processes bytes
			n, err := reader.Read(buf) // loading chunk into buffer
			// fmt.Println(string(buf) + "\n")
			buf = buf[:n]
			if n == 0 {
				if err != nil {
					b.Error(err)
					break
				}
				if err == io.EOF {
					break
				}
			}
		}
	}
}
