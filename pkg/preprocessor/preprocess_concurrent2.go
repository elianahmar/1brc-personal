package preprocessor

import (
	"bufio"
	"bytes"
	"fmt"
	"math"
	"strconv"
	"sync"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type chunk struct {
	bufSize int
	offset  int
	idx     int
}

type readChunk struct {
	buffer []byte
	offset int
	idx    int
}

func ReadFileConcurrent2(path string) map[model.City]*model.Measurement {
	wg := &sync.WaitGroup{}
	fileScanner := bufio.Scanner{}
	defer file.Close()

	fileStats := utils.PanicOnError(file.Stat())
	fileSizeBytes := fileStats.Size()
	chunkSize := 100000 // characters

	goRoutines := fileSizeBytes / int64(chunkSize)

	hasLeftover := fileSizeBytes%int64(chunkSize) > 0
	if hasLeftover {
		goRoutines += 1
	}
	chunks := make([]chunk, goRoutines)
	for i := 0; i < int(goRoutines); i++ {
		chunks[i].bufSize = chunkSize
		chunks[i].offset = i * chunkSize
		chunks[i].idx = i
	}
	readChunks := make([]readChunk, goRoutines)

	wg.Add(int(goRoutines))
	// spawn go routines for reading each chunk
	for i := 0; i < int(goRoutines); i++ {
		go func(i int) {
			defer wg.Done()
			// fmt.Println("reading chunk %d", i)
			chunk := &chunks[i]
			buffer := make([]byte, chunk.bufSize)
			file.ReadAt(buffer, int64(chunk.offset))
			readChunks[i] = readChunk{idx: i, buffer: buffer, offset: chunk.offset}
		}(i)
	}
	wg.Wait()

	reconcileLines(readChunks)
	measurements := collectDataConcurrent(readChunks)

	return measurements
}

// This method is needed to ensure that chunk reading doesn't cut off a line
// If we discover a line to be cut off. Then we will push it up to the previous chunk
func reconcileLines(readChunks []readChunk) {
	// 1. Go through each read chunk sequentially
	// 2. If current chunk doesn't end with newline character, then I need to grab everything from the next chunk up to /n and append it to the current chunk
	for i := 0; i < len(readChunks)-1; i++ {
		reconcileChunks(&readChunks[i], &readChunks[i+1])
	}
}

// TODO: this is completely cooking my runtime
func reconcileChunks(currChunk *readChunk, nextChunk *readChunk) {
	// If we cut off at the right point. Return early
	bufLen := len(currChunk.buffer)
	if currChunk.buffer[bufLen-1] == '\n' {
		return
	}
	// Compute the first '\n' in next chunk and append it to currChunk
	// and remove from nextChunk
	breakPoint := utils.First(nextChunk.buffer, func(b byte) bool { return b == '\n' })
	if breakPoint == -1 {
		return
	}
	currChunk.buffer = append(currChunk.buffer, nextChunk.buffer[:breakPoint+1]...)
	nextChunk.buffer = nextChunk.buffer[breakPoint+1:]
}

func collectDataConcurrent(readChunks []readChunk) map[model.City]*model.Measurement {
	// For now, I'll go through everything sequentially. Once I feel good about implementation
	// I'll make this parallel
	measurements := make(map[model.City]*model.Measurement, 500)
	totalLines := 0
	for _, chunk := range readChunks {
		totalLines += processChunk(chunk, measurements, len(readChunks))
	}
	utils.PanicOnCondition(totalLines != 1000000000, "not all cities processed")
	return measurements
}

func processChunk(chunk readChunk, measurements map[model.City]*model.Measurement, numChunks int) int {
	// Process each byte
	newline := []byte{'\n'}
	processed := 0
	lineSeparated := bytes.Split(chunk.buffer, newline)
	for i := range lineSeparated {
		// line := lineSeparated[i]
		// utils.PanicOnCondition(len(line) <= 0, fmt.Sprintf("line %d/%d is empty... shouldn't happen. Chunk index: %d, total chunks: %d", i, len(lineSeparated), chunk.idx, numChunks))
		// utils.PanicOnCondition(line[len(line)-1] == '\n', "line not processed correctly. Every line should end with new line break")
		city, temp, err := processLineByte(lineSeparated[i])
		if err != nil {
			continue
		}
		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{City: city}
		}
		measurements[city].Temps += temp
		measurements[city].Count += 1
		measurements[city].Max = math.Max(measurements[city].Max, temp)
		measurements[city].Min = math.Min(measurements[city].Min, temp)
		processed += 1
	}
	return processed
}

func processLineByte(bSlice []byte) (model.City, float64, error) {
	semicolon := []byte{';'}
	split := bytes.Split(bSlice, semicolon)
	if len(split) != 2 {
		return "", 0.0, fmt.Errorf("split not long enough")
	}
	// fmt.Println("%v", split)
	// utils.PanicOnCondition(len(split) != 2, "byte slice not containing both city and temp")
	dig := utils.PanicOnError(strconv.ParseFloat(string(split[1]), 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.City(split[0]), temp, nil
}
