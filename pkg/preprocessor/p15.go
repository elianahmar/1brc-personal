package preprocessor

import (
	"os"
	"strconv"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P15 struct {
	Path     string
	ChanSize int
}

func NewP15(path string) *P15 {
	return &P15{
		Path: path,
	}
}

func (p15 *P15) Compute() map[string]*model.MeasurementInt { // 12 seconds.

	// Produce the ranges first
	ranges := files.ChunkFileImproved(p15.Path)
	mChan := make(chan map[string]*model.MeasurementInt, len(ranges))
	wg := sync.WaitGroup{}
	file, _ := os.Open(p15.Path)

	wg.Add(len(ranges))
	// Separate go routines for each range. Each go routine will build a map internally
	// and push it to a channel of maps which are processed on main thread
	for _, r := range ranges {
		go func(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File) {
			defer wg.Done()
			utils.PanicIf(r.Start >= r.End, "bounds are wrong")
			p15.processRange(r, mChan, file)
		}(r, mChan, file)
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
func (p15 *P15) processRange(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File) {
	numByte := make([]byte, 0, 8) // TODO: Sync Pool this?
	delim, period := byte(';'), byte('.')
	semiIndex, temp, iters, nxtLine := 0, 0, 0, 0

	localMeasurement := make(map[string]*model.MeasurementInt, 512)
	buff := make([]byte, r.End-r.Start)
	file.ReadAt(buff, r.Start)
	start := 0
	newline := byte('\n')
	for start < len(buff) {
		// println(start)
		buff = buff[start:]
		numByte = numByte[:0]

		// Parse for city. I'm assuming boundaries are correct
		ptr := 0
		iters = 0
		for buff[ptr] != delim {
			ptr++
			iters++
			utils.PanicIf(iters > 64, "infinite looping city parsing")
		}
		city := unsafe.String(&buff[0], ptr)

		// Parse the number
		iters = 0
		nxtLine = -1
		semiIndex = ptr
		ptr++
		for ptr < len(buff) {
			nb := buff[ptr]
			if nb == newline {
				nxtLine = ptr
				break
			}
			if nb != period {
				numByte = append(numByte, nb)
			}
			ptr++
			iters++
			utils.PanicIf(iters > 24, "infinite looping, num parsing")
		}

		if nxtLine == -1 {
			break
		}

		temp, _ = strconv.Atoi(unsafe.String(&numByte[0], len(numByte)))

		measurement, exists := localMeasurement[city] // Lookup trick. city underlying byte array can change but we can use it for lookup
		if !exists {
			// NOTE: Was casting string to string which doesn't copy. That's why map data was wrong
			cityName := string(buff[:semiIndex])
			measurement = &model.MeasurementInt{City: cityName}
			localMeasurement[cityName] = measurement
		}
		measurement.Temps += temp
		measurement.Count += 1
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
		start = ptr + 1
	}
	mChan <- localMeasurement
}
