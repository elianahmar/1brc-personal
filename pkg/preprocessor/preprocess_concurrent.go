package preprocessor

import (
	"bytes"
	"os"
	"sync"

	"github.com/throwea/1brc-go/pkg/model"
	m "github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ReadFileConcurrent(path string) map[m.City]*m.Measurement {
	wg := &sync.WaitGroup{}
	file := utils.PanicOnError(os.Open(path))
	defer file.Close()

	fileStats := utils.PanicOnError(file.Stat())
	fileSizeBytes := fileStats.Size()
	chunkSize := 100000 // characters

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
	readChunks := make([]m.ReadChunk, goRoutines)

	wg.Add(int(goRoutines))
	// spawn go routines for reading each chunk
	for i := 0; i < int(goRoutines); i++ {
		go func(i int) {
			defer wg.Done()
			// fmt.Println("reading chunk %d", i)
			chunk := &chunks[i]
			buffer := make([]byte, chunk.BufSize)
			file.ReadAt(buffer, int64(chunk.Offset))
			readChunks[i] = m.ReadChunk{Idx: i, Buffer: buffer, Offset: chunk.Offset}
		}(i)
	}
	wg.Wait()

	reconcileLines(readChunks)
	measurements := collectDataConcurrent(readChunks)

	return measurements
}

// This method is needed to ensure that chunk reading doesn't cut off a line
// If we discover a line to be cut off. Then we will push it up to the previous chunk
func reconcileLines(readChunks []m.ReadChunk) {
	// 1. Go through each read chunk sequentially
	// 2. If current chunk doesn't end with newline character, then I need to grab everything from the next chunk up to /n and append it to the current chunk
	for i := 0; i < len(readChunks)-1; i++ {
		reconcileChunks(&readChunks[i], &readChunks[i+1])
	}
}

func reconcileChunks(currChunk *m.ReadChunk, nextChunk *m.ReadChunk) {
	// If we cut off at the right point. Return early
	bufLen := len(currChunk.Buffer)
	if currChunk.Buffer[bufLen-1] == '\n' {
		return
	}
	// Compute the first '\n' in next chunk and append it to currChunk
	// and remove from nextChunk
	breakPoint := utils.First(nextChunk.Buffer, func(b byte) bool { return b == '\n' })
	currChunk.Buffer = append(currChunk.Buffer, nextChunk.Buffer[:breakPoint+1]...)
	nextChunk.Buffer = nextChunk.Buffer[breakPoint+1:]
}

func collectDataConcurrent(readChunks []m.ReadChunk) map[model.City]*model.Measurement {
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

func processChunk(chunk m.ReadChunk, measurements map[model.City]*model.Measurement, numChunks int) int {
	// Process each byte
	newline := []byte{'\n'}
	processed := 0
	lineSeparated := bytes.Split(chunk.Buffer, newline)
	for i := range lineSeparated {
		// line := lineSeparated[i]
		// utils.PanicOnCondition(len(line) <= 0, fmt.Sprintf("line %d/%d is empty... shouldn't happen. Chunk index: %d, total chunks: %d", i, len(lineSeparated), chunk.idx, numChunks))
		// utils.PanicOnCondition(line[len(line)-1] == '\n', "line not processed correctly. Every line should end with new line break")
		city, temp, err := processLineByte(lineSeparated[i])
		if err != nil {
			continue
		}
		updateMeasurements(measurements, city, temp)

		processed += 1
	}
	return processed
}
