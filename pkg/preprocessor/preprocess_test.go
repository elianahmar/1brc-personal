package preprocessor

import (
	"bufio"
	"bytes"
	"io"
	"math"
	"os"
	"strconv"
	"sync"
	"testing"

	"github.com/throwea/1brc-go/pkg/model"
	m "github.com/throwea/1brc-go/pkg/model"
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

func BenchmarkFileScanning(b *testing.B) { // 1.481s (Small Data) 15.690s full dataset
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()
		fileScanner := bufio.NewScanner(file)
		for fileScanner.Scan() {
			fileScanner.Bytes()
			// process the line itself
		}
	}
}

// 1.492s (Small data) 12.582s (Full Dataset) 100000 bytes
// 6.899s (Full Dataset) 4096 * (32) ~= 128 bytes
func BenchmarkFileChunking(b *testing.B) {
	for b.Loop() {
		wg := &sync.WaitGroup{}
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()

		fileStats := utils.PanicE(file.Stat())
		fileSizeBytes := fileStats.Size()
		chunkSize := 4096 * 32 // ~128 mb?

		goRoutines := fileSizeBytes / int64(chunkSize)

		hasLeftover := fileSizeBytes%int64(chunkSize) > 0
		if hasLeftover {
			goRoutines += 1
		}
		chunks := make([]m.Chunk, goRoutines)
		for i := 0; i < int(goRoutines); i++ {
			chunks[i].BufSize = chunkSize
			chunks[i].Offset = i * chunkSize
			chunks[i].Idx = i
		}
		readChunks := make([]*m.ReadChunk, goRoutines)

		wg.Add(int(goRoutines))
		// spawn go routines for reading each chunk
		for i := 0; i < int(goRoutines); i++ {
			go func(i int) {
				defer wg.Done()
				chunk := &chunks[i]
				buffer := make([]byte, chunk.BufSize)
				file.ReadAt(buffer, int64(chunk.Offset))
				readChunks[i] = &m.ReadChunk{Idx: i, Buffer: buffer, Offset: chunk.Offset}
			}(i)
		}
		wg.Wait()

	}
}

func benchmarkUnsafe(b *testing.B) {
	for b.Loop() {
		// 117 seconds. Fastest yet. All single threaded????
		// Brute force this. Read line by line and update a table
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()
		fileScanner := bufio.NewScanner(file)
		delim := []byte{';'}
		measurements := make(map[string]*model.Measurement, 512) // 512 bc it's power of 2
		for fileScanner.Scan() {
			line := fileScanner.Bytes() // NOTE: unsafe is no good here. Per the docs. The underlying array can be overwritten
			// process the line itself
			city, num, found := bytes.Cut(line, delim) // Returns original array. Unsafe is no good here either
			cityName := utils.BytesToString(city)
			utils.PanicIf(!found, "bytes not found?")
			temp := utils.PanicE(strconv.ParseFloat(string(num), 64))
			if _, exists := measurements[cityName]; !exists {
				measurements[cityName] = &model.Measurement{City: cityName}
			}
			measurements[cityName].Temps += temp
			measurements[cityName].Count += 1
			measurements[cityName].Max = math.Max(measurements[cityName].Max, temp)
			measurements[cityName].Min = math.Min(measurements[cityName].Min, temp)
		}
	}
}

// Small data: 1.484s, Full Dataset: 11.931s
func BenchmarkReadFile(b *testing.B) {
	for b.Loop() {
		os.ReadFile("../../../1brc-go/measurements.txt")
	}
}

// Full Dataset: 184 seconds. Will not be splitting
func benchmarkReadFile_Split(b *testing.B) {
	for b.Loop() {
		file, _ := os.ReadFile("../../../1brc-go/measurements.txt")
		bytes.Split(file, []byte{'\n'})
	}
}

// Full Dataset: 17.573 seconds
func BenchmarkBufioReader_ReadSlice(b *testing.B) {
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		fileReader := bufio.NewReader(file)
		for {
			_, err := fileReader.ReadSlice('\n')
			// fmt.Println(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatalf("reader failed: %v", err)
			}
		}
	}
}

// Full Dataset: 31.8 seconds... Makes sense ReadBytes() copies underlying array
func BenchmarkBufioReader_ReadBytes(b *testing.B) {
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		fileReader := bufio.NewReader(file)
		for {
			_, err := fileReader.ReadBytes('\n')
			// fmt.Println(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatalf("reader failed: %v", err)
			}
		}
	}
}

// Full Dataset:
func BenchmarkBufioReader_ReadLine(b *testing.B) {
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		fileReader := bufio.NewReader(file)
		for {
			_, pref, err := fileReader.ReadLine()
			// fmt.Println(n)
			if err == io.EOF {
				break
			}
			if err != nil || pref {
				b.Fatalf("reader failed: %v", err)
			}
		}
	}
}

// NOTE: for setting buffer; param 1 is the buffer size and param two is max token size
// For large files it's ideal to have a large buffer, that way we can read many tokens into the buffer at once
// The "token" in this case, is what we are telling the scanner to split on. Which in the default case is '\n'
//
// 1mb + Full Dataset -> 15.409s
// 8mb + Full Dataset -> 16.3s
// 4mb + Full Dataset -> 15.8s
// 2mb + Full Dataset -> 15.8s
// 2mb + 1mb maxtoken Full Dataset -> 15.0s
func BenchmarkFileScanning_1mbBuffer(b *testing.B) { // 1.481s (Small Data) 15.690s full dataset
	for b.Loop() {
		file := utils.PanicE(os.Open("../../../1brc-go/measurements.txt"))
		defer file.Close()
		mb := 2
		bufSize := mb * 1024 * 1024
		fileScanner := bufio.NewScanner(file)
		fileScanner.Buffer(make([]byte, bufSize), 1024)
		for fileScanner.Scan() {
			fileScanner.Bytes()
			// process the line itself
		}
	}
}
