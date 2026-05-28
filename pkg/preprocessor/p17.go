package preprocessor

import (
	"io"
	"os"
	"sync"
	"unsafe"

	"github.com/throwea/1brc-go/pkg/files"
	"github.com/throwea/1brc-go/pkg/model"
)

const (
	semicolon = byte(';')
	newline   = byte('\n')
	minus     = byte('-')
	zero      = byte('0')
	nine      = byte('9')
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

func (p17 *P17) Compute() map[string]*model.MeasurementInt { // 3.8 seconds

	rChan := make(chan model.Range, 1000) // TODO: make this configurable
	rSig := make(chan bool)
	mChan := make(chan map[string]*model.MeasurementInt, 1000)
	file, _ := os.Open(p17.Path)

	go files.ChunkFileAsync(p17.Path, rChan)

	go func(mChan chan map[string]*model.MeasurementInt, file *os.File) {
		wg := &sync.WaitGroup{}
		for r := range rChan { // We are receiving all of the ranges. I validated with prints. saw 3290
			wg.Add(1)
			go func(r model.Range, mChan chan map[string]*model.MeasurementInt, file *os.File, wg *sync.WaitGroup) {
				p17.processRange(r, mChan, file, wg)
			}(r, mChan, file, wg)
		}
		wg.Wait()
		rSig <- true
	}(mChan, file)

	go func(rSig chan bool) {
		<-rSig
		close(mChan)
	}(rSig)

	finalMeasure := make(map[string]*model.MeasurementInt, 512)
	for localMeasurement := range mChan {
		for city, newMeasure := range localMeasurement {
			measurement, exists := finalMeasure[city]
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
	n, err := file.ReadAt(buff, r.Start)
	if err != nil && err != io.EOF {
		panic(err)
	}
	buff = buff[:n] // NOTE: need to do this because it isn't guaranteed that I'll read in all of the bytes

	ptr := 0
	temp := 0
	for ptr < n {
		// Read up to the ';'
		start := ptr
		for buff[ptr] != semicolon {
			ptr++
		}
		cityEnd := ptr
		city := unsafe.String(&buff[start], ptr-start)
		// Move the ptr forward off the semicolon
		ptr++
		temp = 0
		for ptr < n {
			nb := buff[ptr]
			if nb == newline {
				break
			}
			if zero <= nb && nb <= nine {
				temp = temp*10 + int(nb-zero)
			}
			ptr++
		}
		if buff[cityEnd+1] == minus {
			temp *= -1
		}
		ptr++ // move past newline
		measurement, exists := localMeasurement[city]
		if !exists {
			cityName := string(buff[start:cityEnd])
			measurement = &model.MeasurementInt{City: cityName}
			localMeasurement[cityName] = measurement
		}

		measurement.Temps += temp
		measurement.Count++
		measurement.Max = max(measurement.Max, temp)
		measurement.Min = min(measurement.Min, temp)
	}
	mChan <- localMeasurement
}
