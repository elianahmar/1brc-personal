package preprocessor

import (
	"bufio"
	"bytes"
	"math"
	"os"
	"strconv"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

type P4 struct {
	Path     string
	ChanSize int
}

func NewP4(path string, chansize int) *P4 {
	return &P4{
		Path:     path,
		ChanSize: chansize,
	}
}

// THOUGHTS: To read concurrently, I will need to read each byte individually. However, that will pose another problem
// If I chunk based on bytes, then there is a possibility of some lines being cut off. I would have to resolve those lines
// Let me think about this. I read the entire line by line and create an object for each line. What if read in parallel, rejoin the entire

func (p4 *P4) Compute() map[string]*model.Measurement { // 117 seconds. Fastest yet. All single threaded????
	// Brute force this. Read line by line and update a table
	file := utils.PanicE(os.Open(p4.Path))
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	delim := []byte{';'}
	measurements := make(map[string]*model.Measurement, 512) // 512 bc it's power of 2
	for fileScanner.Scan() {
		line := fileScanner.Bytes()
		// process the line itself
		city, num, found := bytes.Cut(line, delim)
		cityName := string(city)
		utils.PanicIf(!found, "bytes not found?", nil)
		temp := utils.PanicE(strconv.ParseFloat(string(num), 64))
		if _, exists := measurements[cityName]; !exists {
			measurements[cityName] = &model.Measurement{City: cityName}
		}
		measurements[cityName].Temps += temp
		measurements[cityName].Count += 1
		measurements[cityName].Max = math.Max(measurements[cityName].Max, temp)
		measurements[cityName].Min = math.Min(measurements[cityName].Min, temp)
	}
	return measurements
}

// func (p4 *P4) Compute() map[string]*model.Measurement { // Way slower took 199 seconds. Wtf
// 	// Brute force this. Read line by line and update a table
// 	file := utils.PanicE(os.Open(p4.Path))
// 	defer file.Close()
// 	fileScanner := bufio.NewScanner(file)
// 	delim := []byte{';'}
// 	measurements := make(map[string]*model.Measurement, 512) // 512 bc it's power of 2
// 	lineChan := make(chan lineData, 1000000)
// 	wg := sync.WaitGroup{}
//
// 	wg.Add(2)
// 	go func() {
// 		defer wg.Done()
// 		for fileScanner.Scan() {
// 			line := fileScanner.Bytes()
// 			// process the line itself
// 			city, num, found := bytes.Cut(line, delim)
// 			cityName := string(city)
// 			utils.PanicIf(!found, "bytes not found?")
// 			temp := utils.PanicE(strconv.ParseFloat(string(num), 64))
// 			lineChan <- lineData{cityName, temp}
//
// 		}
// 		close(lineChan)
// 	}()
//
// 	go func() {
// 		defer wg.Done()
// 		for lineData := range lineChan {
// 			cityName, temp := lineData.City, lineData.Temp
// 			if _, exists := measurements[cityName]; !exists {
// 				measurements[cityName] = &model.Measurement{City: cityName}
// 			}
// 			measurements[cityName].Temps += temp
// 			measurements[cityName].Count += 1
// 			measurements[cityName].Max = math.Max(measurements[cityName].Max, temp)
// 			measurements[cityName].Min = math.Min(measurements[cityName].Min, temp)
// 		}
// 	}()
// 	wg.Wait()
// 	return measurements
// }
