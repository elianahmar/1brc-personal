package preprocessor

import (
	"os"
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

	readChunks = reconcileLines(readChunks)
	measurements := collectDataConcurrent()

	return measurements
}

// This method is needed to ensure that chunk reading doesn't cut off a line
// If we discover a line to be cut off. Then we will push it up to the previous chunk
func reconcileLines(readChunks []readChunk) []readChunk {
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

func collectDataConcurrent() map[model.City]*model.Measurement {}
