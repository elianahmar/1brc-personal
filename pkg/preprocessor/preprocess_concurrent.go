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
}

func collectDataConcurrent() map[model.City]*model.Measurement {}
