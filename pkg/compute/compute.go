package compute

import (
	"fmt"
	"math"

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

// NOTE: I can just convert to string then simply put a period before smallest digit. Done. Try that if computations don't work
func ComputeAvgInt(measurements map[string]*model.MeasurementInt) map[string]*model.Predicted {
	predictions := make(map[string]*model.Predicted)
	for city, measurement := range measurements {
		avg := math.Round(float64(measurement.Temps) / float64(measurement.Count))
		avg /= 10
		predictions[city] = &model.Predicted{
			Min: utils.TruncateNaive(float64(measurement.Min)/10.0, 0.1),
			Max: utils.TruncateNaive(float64(measurement.Max)/10.0, 0.1),
			Avg: utils.TruncateNaive(math.Round(avg), 0.1),
		}
	}
	return predictions
}
