package preprocessor

import (
	"io"
	"os"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
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

func (p17 *P17) processRange(
	r model.Range,
	mChan chan map[string]*model.MeasurementInt,
	file *os.File,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	localMeasurement := make(map[string]*model.MeasurementInt, 512)

	buff := make([]byte, r.End-r.Start+1)
	n, err := file.ReadAt(buff, r.Start)
	if err != nil && err != io.EOF {
		panic(err)
	}
	buff = buff[:n] // TODO: why do I need this. COmmenting this out fails my solution

	const (
		semicolon = byte(';')
		newline   = byte('\n')
		period    = byte('.')
		minus     = byte('-')
		zero      = byte('0')
		nine      = byte('9')
	)

	N := len(buff)
	ptr := 0
	/// Parser code /////
	for ptr < N {
		// Read up to the ';'
		start := ptr
		for buff[ptr] != semicolon {
			ptr++
		}
		cityEnd := ptr
		city := unsafe.String(&buff[start], ptr-start)
		// Move the ptr forward off the semicolon
		ptr++
		isNeg := buff[ptr] == minus
		ptr++
		temp := 0
		for ptr < N {
			nb := buff[ptr]
			if nb == newline {
				break
			}
			if zero <= nb && nb <= nine {
				temp = temp*10 + int(nb-zero)
			}
			ptr++
		}
		// Flip sign if needed
		if isNeg {
			temp *= -1
		}

		if ptr >= N {
			break
		}

		ptr++ // move past newline

		/// Parser code /////

		// 3. Update local map
		measurement, exists := localMeasurement[city]
		if !exists {
			cityName := string(buff[start:cityEnd])
			// println(cityName, temp)
			measurement = &model.MeasurementInt{City: cityName}
			localMeasurement[cityName] = measurement
		}
		// utils.PanicIf(temp == 0.0, "temp not parsed correctly", nil)

		measurement.Temps += temp
		measurement.Count++
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	mChan <- localMeasurement
}
