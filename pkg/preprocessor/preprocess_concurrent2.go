package preprocessor

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ReadFileConcurrent2(path string) map[model.City]*model.Measurement {
	wg := &sync.WaitGroup{}
	readFileStart := time.Now()
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
	chunks := make([]chunk, goRoutines)
	for i := 0; i < int(goRoutines); i++ {
		chunks[i].bufSize = chunkSize
		chunks[i].offset = i * chunkSize
		chunks[i].idx = i
	}
	readChunks := make([]*readChunk, goRoutines)

	wg.Add(int(goRoutines))
	// spawn go routines for reading each chunk
	for i := 0; i < int(goRoutines); i++ {
		go func(i int) {
			defer wg.Done()
			// fmt.Println("reading chunk %d", i)
			chunk := &chunks[i]
			buffer := make([]byte, chunk.bufSize)
			file.ReadAt(buffer, int64(chunk.offset))
			readChunks[i] = &readChunk{idx: i, buffer: buffer, offset: chunk.offset}
		}(i)
	}
	wg.Wait()
	fmt.Println("time taken to process all the bytes %d", time.Since(readFileStart).Seconds())

	cutLinesConcurrent(readChunks)
	// reconcileLines2(readChunks) // TODO: This is killing me
	// measurements := collectDataConcurrent(readChunks)

	return measurements
}

type Line struct {
	// What chunk it appears in
	chunkIdx int
	// Full line as byte slice
	line []byte
	// Index of the line after we split the bytes on '\n'
	lineIdx int
}

// What is the idea here? I have multiple chunks of bytes that I have to reconcile somehow. At the boundaries they will be cutoff
// So I have two cases I have to deal with
// So let's do this. First and last line go to merge chan. One edge case we need to deal with is if chunk == 0 and lineidx == 0 then we skip it or if it's the last chunk and last line. Since we can guarantee it's a valid line
// For the rest, I'm still not clear how I will connect them all concurrently.
func cutLinesConcurrent(readChunks []*readChunk) {
	mergeChan := make(chan Line, len(readChunks)-1)
	fullLineChan := make(chan Line, len(readChunks)-1)
	newline := []byte{'\n'}
	wg := &sync.WaitGroup{}
	ops := atomic.Uint64{}

	wg.Add(3)
	// Producer
	go func() {
		defer wg.Done()
		for _, chunk := range readChunks { // TODO: I could even split this up using go routines
			// TODO: come back to the merge line case in a bit
			splitLines := bytes.Split(chunk.buffer, newline)
			linesToProcess := len(splitLines)
			// Push the very first and very last line
			mergeChan <- Line{chunkIdx: chunk.idx, line: splitLines[0], lineIdx: 0}
			mergeChan <- Line{chunkIdx: chunk.idx, line: splitLines[linesToProcess-1], lineIdx: linesToProcess}

			for i := 1; i < linesToProcess-1; i++ {
				fullLineChan <- Line{chunkIdx: chunk.idx, line: splitLines[i], lineIdx: i}
			}
		}
	}()

	// Consumer 1 for good lines
	go func() {
		defer wg.Done()
		for goodLine := range fullLineChan {
			ops.Add(1)
		}
	}()

	// Consumer 2 for bad lines
	go func() {
		defer wg.Done()
	}()
	wg.Wait()
	linesRead := ops.Load()
	utils.PanicOnCondition(linesRead != 1000000000, fmt.Sprintf("did not process all lines. Lines read = %d", linesRead))
}
