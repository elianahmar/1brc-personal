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

// NOTE: I can just convert to string then simply put a period before smallest digit. Done. Try that if computations don't work
//
//	func ComputeAvgInt(measurements map[string]*model.MeasurementInt) map[string]*model.Predicted {
//		predictions := make(map[string]*model.Predicted)
//		for city, measurement := range measurements {
//			avg := math.Round(float64(measurement.Temps) / float64(measurement.Count))
//			avg /= 10
//			predictions[city] = &model.Predicted{
//				Min: utils.TruncateNaive(float64(measurement.Min)/10.0, 0.1),
//				Max: utils.TruncateNaive(float64(measurement.Max)/10.0, 0.1),
//				Avg: utils.TruncateNaive(math.Round(avg), 0.1),
//			}
//		}
//		return predictions
//	}

// NOTE: If temps is positive let's ceil the Temps/Avg, else floor it
func ComputeAvgStrConv(measurements map[string]*model.MeasurementInt) map[string]*model.Predicted {
	predictions := make(map[string]*model.Predicted)
	fmt.Println(printFullMeasurements(measurements))
	for city, measurement := range measurements {
		//
		avg := measurement.Temps / measurement.Count
		predictions[city] = &model.Predicted{
			City: city,
			Min:  convertToStr(measurement.Min),
			Max:  convertToStr(measurement.Max),
			Avg:  convertToStr(avg),
		}
	}
	return predictions
}

func convertToStr(num int) string {
	res := strings.Builder{}
	numStr := strconv.Itoa(num)
	N := len(numStr)

	utils.PanicIf(N > 5, "") // Temp should never exceed 5 bytes (ex. -45.4)
	for i := range N {
		if i == N-1 {
			res.WriteByte('.')
		}
		res.WriteByte(numStr[i])
	}
	return res.String()
}

func printFullMeasurements(measurements map[string]*model.MeasurementInt) string {
	res := strings.Builder{}
	for city, mment := range measurements {
		line := fmt.Sprintf("City = %s, Min = %d, Max = %d, Count = %d, Temps = %d\n", city, mment.Min, mment.Max, mment.Count, mment.Temps)
		res.WriteString(line)
	}
	return res.String()
}
