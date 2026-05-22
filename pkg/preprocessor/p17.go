package preprocessor

import (
	"fmt"
	"os"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P17 struct {
	Path     string
	ChanSize int
}

func NewP17(path string) *P17 {
	return &P17{
		Path: path,
	}
}

// WARN: the range isn't inclusive... that's why + 1 is needed on the buffer
// Also, the very last line always includes a newline break
// So for a single pass parse, I need to scan the whole line and if their are no more bytes
// Just break.
func readRange(r model.Range, path string) {
	input := utils.PanicE(os.Open(path))
	defer input.Close()
	buff := make([]byte, r.End-r.Start+1)
	input.ReadAt(buff, r.Start)

	newFilePath := "./firstrange.txt"
	newFile, err := os.Create(newFilePath)
	if err != nil {
		panic("Failed to create file: " + err.Error())
	}
	defer newFile.Close()

	length, err := newFile.Write(buff)
	if err != nil {
		panic(err)
	}
	fmt.Printf("File name: %s\n", newFile.Name())
	fmt.Printf("File length: %d bytes\n", length)
	return
}

func (p17 *P17) Compute() map[string]*model.MeasurementInt { // 4.5 seconds.

	rChan := make(chan model.Range, 1000) // TODO: make this configurable
	rSig := make(chan bool)
	mChan := make(chan map[string]*model.MeasurementInt, 1000)
	file, _ := os.Open(p17.Path)

	go files.ChunkFileAsync(p17.Path, rChan)
	// Separate go routines for each range. Each go routine will build a map internally
	// and push it to a channel of maps which are processed on main thread
	go func(mChan chan map[string]*model.MeasurementInt, file *os.File) {
		wg := &sync.WaitGroup{}
		for r := range rChan { // We are receiving all of the ranges. I validated with prints. saw 3290
			wg.Add(1)
			go func(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File, wg *sync.WaitGroup) {
				p17.processRange(r, mChan, file, wg)
			}(r, mChan, file, wg)
		}
		// BUG: I see the issue I think. We receive all of the ranges and spawn go routines
		// To process them, but we aren't actually done, until all of those go routines complete
		// I could cheat since I know we have to process 3290. But I want to implement this with the assumption that
		// I don't know how many ranges appear. That will allow me to tune different params like the chunk size
		//
		// BUG:Hack solution, ChunkFileAsync can push the total number of ranges to a chan
		// In the main thread, I can count if we have received all ranges and push true to rSig
		// Process Range knows how many times
		//
		// NOTE: fix was simple. just add to waitgroup on every receive from rChan...
		wg.Wait()
		rSig <- true
	}(mChan, file)
	// Spawn another go routine which waits for all ranges to be processed and closes
	// the channel so localMeasurement := range mChan can exit after it drains the channel
	go func(rSig chan bool) {
		// TODO: how do I synchronize in this case?
		// I'll have the first go routine producing the ranges and pushing to rChan
		// The second go routine will be consuming from range chan
		// Then lastly, on the main thread, I'm receiving from mChan
		//
		// 1. So I think the first thing I'll need to do is spawn a go routine for producing the
		// the ranges. Inside that function we close the channel after we are done
		// 2. Then the second go routine will be for processing the ranges. The second go routine
		// is processing the range and pushing the map to the mChan
		// 3. On the main thread I need to consume from the mChan until it's closed
		//
		//PERF: I can create a separate channel which will block until it receives a value... Easy
		//
		//rChan -> mChan -> main
		<-rSig
		close(mChan)
	}(rSig)

	finalMeasure := make(map[string]*model.MeasurementInt, 512)
	for localMeasurement := range mChan {
		for city, newMeasure := range localMeasurement {
			measurement, exists := finalMeasure[city]
			// IDEA: could we avoid allocating a new object if !exists and just assign
			if !exists {
				measurement = newMeasure
				finalMeasure[city] = measurement
				continue
			}
			measurement.Temps += newMeasure.Temps
			measurement.Count += newMeasure.Count
			measurement.Max = max(measurement.Max, newMeasure.Max)
			measurement.Min = min(measurement.Min, newMeasure.Min)
		}
	}
	return finalMeasure
}

func (p17 *P17) processRange(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File, wg *sync.WaitGroup) {
	defer wg.Done()

	localMeasurement := make(map[string]*model.MeasurementInt, 512)
	buff := make([]byte, r.End-r.Start+1)
	file.ReadAt(buff, r.Start)

	utils.PanicIf(buff[0] == byte('\n'), fmt.Sprintf("not starting at new line, chunk number = %d", r.Index), nil)
	if buff[len(buff)-1] != byte('\n') {
		readRange(r, p17.Path)
		panic(fmt.Sprintf("not ending at new line, chunk number = %d", r.Index))
	}

	// NOTE: Based on these asserts I can guarantee we are creating the chunks correctly
	// So if every buffer starts at a character and ends with a newline, then we should be checking every index up to the end
	// So if we have 10 bytes in the buffer then our pointer needs to [0, 9]

	N, start := len(buff), 0
	ptr := 0
	for ptr < N {
		start = ptr
		temp, city, nlIdx, dlIdx, nlFound := ParseLine(ptr, buff, N)
		if !nlFound {
			break
		}
		ptr = nlIdx + 1

		// fmt.Println(fmt.Sprintf("%s, [%d, %d]", city, start, dlIdx))
		measurement, exists := localMeasurement[city] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			// NOTE: Was casting string to string which doesn't copy. That's why map data was wrong
			cityName := string(buff[start:dlIdx])
			measurement = &model.MeasurementInt{City: cityName}
			localMeasurement[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	mChan <- localMeasurement
}

func ParseLine(start int, buff []byte, N int) (int, string, int, int, bool) {
	var (
		delim                = byte(';')
		newline              = byte('\n')
		zero, nine, negative = byte('0'), byte('9'), byte('-')
		temp, delimIdx       = 0, 0
		newLineFound         = false
	)

	ptr := start
	utils.PanicIf(buff[start] == newline, "should not be starting a newline", nil)
	for ptr <= N {
		if buff[ptr] == delim {
			delimIdx = ptr - 1
			break
		}
		ptr++
	}

	ptr++ // move past the ';'
	temp = 0
	isNeg := buff[ptr] == negative
	for ptr <= N {
		nb := buff[ptr]
		if nb == newline {
			newLineFound = true
			break
		}
		if zero <= nb && nb <= nine {
			temp = (temp * 10) + int(nb-zero)
		}
		ptr++
	}
	ptr++                                               // So that we move past the newline break
	city := unsafe.String(&buff[start], delimIdx-start) // BUG: This is printing the full line
	if isNeg {
		temp *= -1
	}
	return temp, city, ptr, delimIdx, newLineFound
}
