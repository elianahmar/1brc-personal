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

func ComputeAvgInt(measurements map[string]*model.MeasurementInt) {
	for _, measurement := range measurements {
		avg := float64(measurement.Temps) / float64(measurement.Count)
		avg /= 10
		measurement.Avg = utils.TruncateNaive(math.Round(avg), 0.1)
	}
}
