package preprocessor

import (
	"bytes"
	"fmt"
	"math"
	"strconv"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func processLineByte(bSlice []byte) (model.City, float64, error) {
	semicolon := []byte{';'}
	split := bytes.Split(bSlice, semicolon)
	if len(split) != 2 {
		return "", 0.0, fmt.Errorf("split not long enough")
	}
	// fmt.Println("%v", split)
	// utils.PanicOnCondition(len(split) != 2, "byte slice not containing both city and temp")
	dig := utils.PanicOnError(strconv.ParseFloat(string(split[1]), 64))
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.City(split[0]), temp, nil
}

func updateMeasurements(measurements map[model.City]*model.Measurement, city model.City, temp float64) {
	if _, exists := measurements[city]; !exists {
		measurements[city] = &model.Measurement{City: city}
	}
	measurements[city].Temps += temp
	measurements[city].Count += 1
	measurements[city].Max = math.Max(measurements[city].Max, temp)
	measurements[city].Min = math.Min(measurements[city].Min, temp)
}
