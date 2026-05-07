package preprocessor

import (
	"bytes"
	"math"
	"os"
	"strconv"
	"sync"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type chunk struct {
	bufSize int
	offset  int
}

type readChunk struct {
	buffer []byte
	offset int
}

func ReadFileConcurrent(path string, chanSize int) map[model.City]*model.Measurement {
	wg := &sync.WaitGroup{}
	file := utils.PanicOnError(os.Open(path))
	defer file.Close()

	fileStats := utils.PanicOnError(file.Stat())
	fileSizeBytes := fileStats.Size()
	chunkSize := 500 // characters

	goRoutines := fileSizeBytes / int64(chunkSize)

	hasLeftover := fileSizeBytes%int64(chunkSize) > 0
	if hasLeftover {
		goRoutines += 1
	}
	chunks := make([]chunk, goRoutines)
	for i := 0; i < int(goRoutines); i++ {
		chunks[i].bufSize = chunkSize
		chunks[i].offset = i * chunkSize
	}
	readChunks := make([]readChunk, goRoutines)

	wg.Add(int(goRoutines))
	// spawn go routines for reading each chunk
	for i := 0; i < int(goRoutines); i++ {
		go func(i int) {
			defer wg.Done()
			chunk := &chunks[i]
			buffer := make([]byte, chunk.bufSize)
			file.ReadAt(buffer, int64(chunk.offset))
			readChunks[i] = readChunk{buffer: buffer, offset: chunk.offset}
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

func reconcileChunks(currChunk *readChunk, nextChunk *readChunk) {
	// If we cut off at the right point. Return early
	bufLen := len(currChunk.buffer)
	if currChunk.buffer[bufLen] == '\n' {
		return
	}
	// Compute the first '\n' in next chunk and append it to currChunk
	// and remove from nextChunk
	breakPoint := utils.First(nextChunk.buffer, func(b byte) bool { return b == '\n' })
	currChunk.buffer = append(currChunk.buffer, nextChunk.buffer[:breakPoint+1]...)
	nextChunk.buffer = nextChunk.buffer[breakPoint+1:]
}

func collectDataConcurrent(readChunks []readChunk) map[model.City]*model.Measurement {
	// For now, I'll go through everything sequentially. Once I feel good about implementation
	// I'll make this parallel
	var measurements map[model.City]*model.Measurement
	for _, chunk := range readChunks {
		processChunk(chunk, measurements)
	}
	return measurements
}

func processChunk(chunk readChunk, measurements map[model.City]*model.Measurement) {
	// Process each byte
	newline := []byte{'\n'}
	lineSeparated := bytes.Split(chunk.buffer, newline)
	for i := range lineSeparated {
		line := lineSeparated[i]

		utils.PanicOnCondition(line[len(line)-1] == '\n', "line not processed correctly. Every line should end with new line break")
		city, temp := processLineByte(lineSeparated[i])
		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{City: city}
		}
		measurements[city].Temps += temp
		measurements[city].Count += 1
		measurements[city].Max = math.Max(measurements[city].Max, temp)
		measurements[city].Min = math.Min(measurements[city].Min, temp)
	}
}

func processLineByte(bSlice []byte) (model.City, float64) {
	semicolon := []byte{';'}
	split := bytes.Split(bSlice, semicolon)
	utils.PanicOnCondition(len(split) == 2, "byte slice not containing both city and temp")
	dig := utils.PanicOnError(strconv.ParseFloat(string(split[1]), 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.City(split[0]), temp
}
