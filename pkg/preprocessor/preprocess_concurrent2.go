package preprocessor

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/throwea/1brc-go/pkg/model"
	m "github.com/throwea/1brc-go/pkg/model"
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
			// fmt.Println("reading chunk %d", i)
			chunk := &chunks[i]
			buffer := make([]byte, chunk.BufSize)
			file.ReadAt(buffer, int64(chunk.Offset))
			readChunks[i] = &m.ReadChunk{Idx: i, Buffer: buffer, Offset: chunk.Offset}
		}(i)
	}
	wg.Wait()
	fmt.Printf("time taken to process all the bytes %f\n", time.Since(readFileStart).Seconds())

	measurements := cutLinesConcurrent(readChunks)
	return measurements
}

// What is the idea here? I have multiple chunks of bytes that I have to reconcile somehow. At the boundaries they will be cutoff
// So I have two cases I have to deal with
// So let's do this. First and last line go to merge chan. One edge case we need to deal with is if chunk == 0 and lineidx == 0 then we skip it or if it's the last chunk and last line. Since we can guarantee it's a valid line
// For the rest, I'm still not clear how I will connect them all concurrently.
func cutLinesConcurrent(readChunks []*m.ReadChunk) map[model.City]*model.Measurement {
	var (
		mergeChan    = make(chan m.Line, len(readChunks)-1)
		fullLineChan = make(chan m.Line, len(readChunks)-1)
		wg           = &sync.WaitGroup{}
		ops          = &atomic.Uint64{}
		mu           = &sync.Mutex{}
	)

	measurements := make(map[model.City]*model.Measurement)
	totalChunks := len(readChunks)
	wg.Add(3)
	// Producer
	go processChunks(wg, readChunks, mergeChan, fullLineChan)
	// Consumer 1 for good lines. I think here I can have multiple go routines, processing Do that later though because I will need some more synchronization (i.e. mutex or atomics)
	go processMergeChan2(wg, fullLineChan, mergeChan, totalChunks)
	go consumeFullLines(wg, fullLineChan, measurements, ops, mu)

	fmt.Println("all go routines running")
	wg.Wait()
	fmt.Println("all lines processed. Time to calculate the answers")
	// linesRead := ops.Load()
	// utils.PanicOnCondition(linesRead != 1000000000, fmt.Sprintf("did not process all lines. Lines read = %d", linesRead))
	return measurements
}

func consumeFullLines(wg *sync.WaitGroup, fullLineChan chan m.Line, measurements map[model.City]*model.Measurement, ops *atomic.Uint64, mu *sync.Mutex) {
	defer wg.Done()
	workers := 10
	wg.Add(workers)
	for range workers {
		go func(fullLineChan chan m.Line) {
			defer wg.Done()
			for goodLine := range fullLineChan {
				ops.Add(1)
				// line := lineSeparated[i]
				// utils.PanicOnCondition(len(line) <= 0, fmt.Sprintf("line %d/%d is empty... shouldn't happen. Chunk index: %d, total chunks: %d", i, len(lineSeparated), chunk.idx, numChunks))
				// utils.PanicOnCondition(line[len(line)-1] == '\n', "line not processed correctly. Every line should end with new line break")
				city, temp, err := processLineByte(goodLine)
				if err != nil {
					continue
				}
				mu.Lock()
				UpdateMeasurement(measurements, city, temp)
				mu.Unlock()
			}
		}(fullLineChan)
	}
}

func processMergeChan2(wg *sync.WaitGroup, fullLineChan chan m.Line, mergeChan chan m.Line, totalChunks int) {
	defer wg.Done()
	lineMap := make(map[[2]int]m.Line)
	for mergeLine := range mergeChan {
		// Loc[0] = chunk idx, Loc[1] = line index
		cIdx, lIdx := mergeLine.ChunkIdx, mergeLine.LineIdx
		if (cIdx == 0 && lIdx == 0) || (cIdx == totalChunks-1 && lIdx > 0) {
			fullLineChan <- mergeLine
			continue
		}
		// If the line we just received is beginning of a chunk. Put it in map and continue
		beginning := [2]int{cIdx, lIdx}
		if lIdx == 0 {
			lineMap[beginning] = mergeLine
			continue
		}
		ending := [2]int{cIdx + 1, 0}
		otherLine, exists := lineMap[ending]
		if !exists {
			lineMap[beginning] = mergeLine
			mergeChan <- mergeLine // PERF: not sure if I can do this. Essentially requeuing the line until we find it's partner
			continue
		}
		mergedBuffer := slices.Concat(mergeLine.Line, otherLine.Line)

		// fmt.Println(fmt.Sprintf("\nmerged Buffer: %s\n", string(mergedBuffer)))
		newLine := m.Line{ChunkIdx: cIdx, Line: mergedBuffer, LineIdx: lIdx}
		delete(lineMap, beginning)
		// delete(lineMap, beginning)
		fullLineChan <- newLine
	}
	close(fullLineChan)
}

func processChunks(wg *sync.WaitGroup, readChunks []*m.ReadChunk, mergeChan, fullLineChan chan m.Line) {
	defer wg.Done()
	newline := []byte{'\n'}
	for _, chunk := range readChunks { // TODO: I could even split this up using go routines
		// TODO: come back to the merge line case in a bit
		splitLines := bytes.Split(chunk.Buffer, newline)
		linesToProcess := len(splitLines)
		// Push the very first and very last line
		mergeChan <- m.Line{ChunkIdx: chunk.Idx, Line: splitLines[0], LineIdx: 0}
		mergeChan <- m.Line{ChunkIdx: chunk.Idx, Line: splitLines[linesToProcess-1], LineIdx: linesToProcess}

		// NOTE: it's guaranteed that anything between first and last line will be a good line
		for i := 1; i < linesToProcess-1; i++ {
			fullLineChan <- m.Line{ChunkIdx: chunk.Idx, Line: splitLines[i], LineIdx: i}
		}
	}
	close(mergeChan)
}
