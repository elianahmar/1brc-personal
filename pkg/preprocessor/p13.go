package preprocessor

import (
	"bytes"
	"os"
	"strconv"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P13 struct {
	Path     string
	ChanSize int
}

func NewP13(path string) *P13 {
	return &P13{
		Path: path,
	}
}

func (p13 *P13) Compute() map[string]*model.MeasurementInt { // 12 seconds.

	// Produce find the ranges first
	ranges := files.ChunkFileImproved(p13.Path)
	mChan := make(chan map[string]*model.MeasurementInt, len(ranges))
	wg := sync.WaitGroup{}

	wg.Add(len(ranges))
	// Separate go routines for each range. Each go routine will build a map internally
	// and push it to a channel of maps which are processed on main thread
	for _, r := range ranges {
		go func(r model.Range, mChan chan map[string]*model.MeasurementInt) {
			defer wg.Done()
			p13.processRange(r, mChan)
		}(r, mChan)
	}
	// Spawn another go routine which waits for all ranges to be processed and closes
	// the channel so localMeasurement := range mChan can exit after it drains the channel
	go func() {
		wg.Wait()
		close(mChan)
	}()

	finalMeasure := make(map[string]*model.MeasurementInt, 512)
	for localMeasurement := range mChan {
		for city, newMeasure := range localMeasurement {
			measurement, exists := finalMeasure[city]
			if !exists {
				measurement = &model.MeasurementInt{City: city}
				finalMeasure[city] = measurement
			}
			measurement.Temps += newMeasure.Temps
			measurement.Count += newMeasure.Count
			measurement.Max = max(measurement.Max, newMeasure.Max)
			measurement.Min = min(measurement.Min, newMeasure.Min)
		}
	}
	return finalMeasure
}

// TODO: if this is slow don't tie this to the object
func (p13 *P13) processRange(r model.Range, mChan chan map[string]*model.MeasurementInt) {
	numByte := make([]byte, 0, 8) // TODO: Sync Pool this?
	delim, period := byte(';'), byte('.')
	L, N, temp := 0, 0, 0

	// NOTE: Inlining the function doesn't improve speed. I think compiler is probably doing it for me
	parse := func(line []byte) (int, int) {
		numByte = numByte[:0] // clear the array
		L, N = 0, len(line)
		for line[L] != delim {
			L += 1
		}
		delimIdx := L
		L += 1
		for L < N {
			nb := line[L]
			if nb != period {
				numByte = append(numByte, nb)
			}
			L += 1
		}
		temp, _ = strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))
		return temp, delimIdx
	}

	// TODO: can I just pass a single ref to this file object
	file := utils.PanicE(os.Open(p13.Path))
	defer file.Close()

	localMeasurement := make(map[string]*model.MeasurementInt, 512)
	buff := make([]byte, r.End-r.Start+1)
	file.ReadAt(buff, r.Start)
	start := 0
	newline := byte('\n')
	for start <= len(buff) {
		buff = buff[start:]
		nextNL := bytes.IndexByte(buff, newline)
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
