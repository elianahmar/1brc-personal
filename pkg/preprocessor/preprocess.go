package preprocessor

import (
	"fmt"
	"strings"

	"github.com/throwea/1brc-go/pkg/model"
)

func collectData(data) {
	measurements := make(map[model.City]model.Measurement)
	for text := range data {
		measurement := processLine(text)
		split := strings.Split(text, ";")
		city := model.City(split[0])
		if _, exists := measurements[city]; !exists {
			measurements[city] = model.Measurement{}
		}
		measurements[city].Temp += measurement.temp
		measurements[city].Count += 1
		fmt.Printf("%v\n", measurement)
	}
}

// }(data)
