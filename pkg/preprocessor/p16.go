package preprocessor

import (
	"os"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
)

type P16 struct {
	Path     string
	ChanSize int
}

func NewP16(path string) *P16 {
	return &P16{
		Path: path,
	}
}

func (p16 *P16) Compute() map[string]*model.MeasurementInt { // 4.5 seconds.

	rChan := make(chan model.Range, 1000) // TODO: make this configurable
	rSig := make(chan bool)
	mChan := make(chan map[string]*model.MeasurementInt, 1000)
	file, _ := os.Open(p16.Path)

	go files.ChunkFileAsync(p16.Path, rChan)
	// Separate go routines for each range. Each go routine will build a map internally
	// and push it to a channel of maps which are processed on main thread
	go func(mChan chan map[string]*model.MeasurementInt, file *os.File) {
		wg := &sync.WaitGroup{}
		for r := range rChan { // We are receiving all of the ranges. I validated with prints. saw 3290
			wg.Add(1)
			go func(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File, wg *sync.WaitGroup) {
				p16.processRange(r, mChan, file, wg)
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

func (p16 *P16) processRange(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File, wg *sync.WaitGroup) {
	var (
		delim                = byte(';')
		zero, nine, negative = byte('0'), byte('9'), byte('-')
		L, N, temp           = 0, 0, 0
	)
	defer wg.Done()

	// NOTE: Inlining the function doesn't improve speed. I think compiler is probably doing it for me
	parse := func(line []byte) (int, int) {
		L, N = 0, len(line)
		for line[L] != delim {
			L++
		}
		delimIdx := L
		L++
		temp = 0
		isNeg := line[L] == negative
		for L < N {
			nb := line[L]
			if zero <= nb && nb <= nine {
				temp = (temp * 10) + int(nb-zero)
			}
			L++
		}
		if isNeg {
			temp *= -1
		}
		return temp, delimIdx
	}

	localMeasurement := make(map[string]*model.MeasurementInt, 512)
	buff := make([]byte, r.End-r.Start+1)
	file.ReadAt(buff, r.Start)
	start := 0
	newline := byte('\n')
	for start <= len(buff) {
		buff = buff[start:]
		buffLen := len(buff)
		nextNL := -1
		ptr := 0
		for ptr < buffLen {
			if buff[ptr] == newline {
				nextNL = ptr
				break
			}
			ptr++
		}
		if nextNL == -1 {
			break
		}

		temp, delimIdx := parse(buff[:nextNL])
		measurement, exists := localMeasurement[unsafe.String(&buff[0], delimIdx)] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			// NOTE: Was casting string to string which doesn't copy. That's why map data was wrong
			cityName := string(buff[0:delimIdx])
			measurement = &model.MeasurementInt{City: cityName}
			localMeasurement[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
		start = nextNL + 1
	}
	mChan <- localMeasurement
}
