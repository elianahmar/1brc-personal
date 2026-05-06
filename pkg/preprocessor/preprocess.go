package preprocessor

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ReadFile(path string) io.ReadCloser {
	content, err := os.ReadFile(path)
	utils.FatalError(err)
}

func CollectData(data chan string) map[model.City]*model.Measurement {
	measurements := make(map[model.City]*model.Measurement)
	for text := range data {
		measurement := ProcessLine(text)
		split := strings.Split(text, ";")
		city := model.City(split[0])
		if _, exists := measurements[city]; !exists {
			measurements[city] = &model.Measurement{}
		}
		measurements[city].Temps += measurement.Temps
		measurements[city].Count += 1
		fmt.Printf("%v\n", measurement)
	}
	return measurements
}

func ProcessLine(text string) model.Measurement {
	split := strings.Split(text, ";")
	dig, err := strconv.ParseFloat(split[1], 64)
	if err != nil {
		panic(err)
	}
	temp := utils.TruncateNaive(dig, 0.1) // No good. We don't need this much precision
	return model.Measurement{
		City:  split[0],
		Temps: temp,
	}
}
