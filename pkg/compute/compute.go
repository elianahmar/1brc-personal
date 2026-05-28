package compute

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ComputeAvg(measurements map[string]*model.Measurement) {
	// TODO: min, max, avg. Min and Max can be computed as we process
	for _, measurement := range measurements {
		measurement.Avg = utils.TruncateNaive(measurement.Temps/measurement.Count, 0.1)
	}
	fmt.Println("Computed the averages. Time to validate")
}

// NOTE: If temps is positive let's ceil the Temps/Avg, else floor it
func ComputeAvgStrConv(measurements map[string]*model.MeasurementInt) map[string]*model.Predicted {
	predictions := make(map[string]*model.Predicted)
	for city, measurement := range measurements {
		// NOTE: for final answers that have 0 as significand. I need to prepend zero to that result
		avg := float64(measurement.Temps) / float64(measurement.Count) / 10.0
		predictions[city] = &model.Predicted{
			City: city,
			Min:  convertToStr(measurement.Min),
			Max:  convertToStr(measurement.Max),
			Avg:  fmt.Sprintf("%.1f", avg),
		}
	}
	return predictions
}

func convertToStr(num int) string {
	res := strings.Builder{}
	numStr := strconv.Itoa(num)
	N := len(numStr)

	utils.PanicIf(N > 5, "", nil) // Temp should never exceed 5 bytes (ex. -45.4)
	for i := range N {
		if i == N-1 {
			res.WriteByte('.')
		}
		res.WriteByte(numStr[i])
	}
	return res.String()
}
