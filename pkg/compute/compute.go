package compute

import (
	"fmt"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ComputeAvg(measurements map[string]*model.Measurement) {
	// TODO: min, max, avg. Min and Max can be computed as we process
	for city, measurement := range measurements {
		measurements[city].Avg = utils.TruncateNaive(measurement.Temps/measurement.Count, 0.1)
	}

	fmt.Println("Computed the averages. Time to validate")
}
