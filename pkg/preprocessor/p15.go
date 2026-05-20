package preprocessor

import (
	"os"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
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
	var (
		delim                = byte(';')
		zero, nine, negative = byte('0'), byte('9'), byte('-')
		L, N, temp           = 0, 0, 0
	)

	// NOTE: Inlining the function doesn't improve speed. I think compiler is probably doing it for me
	parse := func(line []byte) (int, int) {
		L, N = 0, len(line)
		for line[L] != delim {
			L += 1
		}
		delimIdx := L
		L += 1
		temp = 0
		isNeg := line[L] == negative
		for L < N {
			nb := line[L]
			isChar := zero <= nb && nb <= nine
			if isChar {
				temp *= 10
				temp += int(nb - zero)
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
		nextNL := -1 // This is taking a lot of time
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
