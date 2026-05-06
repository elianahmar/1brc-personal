package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ValidateCorrectness(measurements map[model.City]*model.Measurement) {
	var validation map[model.City]string
	content := utils.PanicOnError(os.ReadFile("./validation.json"))

	utils.PanicOnError(struct{}{}, json.Unmarshal(content, &validation))

	for city, temps := range validation { // NOTE: don't think this is right?
		parsedMin, parsedAvg, parsedMax := convertTemperatures(temps)

		predicted, exists := measurements[city]
		utils.PanicOnCondition(!exists, fmt.Sprintf("no data for city: %s", city))

		errs := validateNumbers(predicted, parsedMin, parsedAvg, parsedMax)
		utils.PanicOnCondition(len(errs) > 0, collectErrs(errs))
	}
}

func convertTemperatures(temps string) (float64, float64, float64) {
	values := strings.Split(temps, "/")
	minActual := utils.PanicOnError(strconv.ParseFloat(values[0], 32))
	avgActual := utils.PanicOnError(strconv.ParseFloat(values[1], 32))
	maxActual := utils.PanicOnError(strconv.ParseFloat(values[2], 32))

	parsedMin := utils.TruncateNaive(minActual, 0.1)
	parsedAvg := utils.TruncateNaive(avgActual, 0.1)
	parsedMax := utils.TruncateNaive(maxActual, 0.1)
	return parsedMin, parsedAvg, parsedMax
}

func validateNumbers(predicted *model.Measurement, parsedMin, parsedAvg, parsedMax float64) []error {
	errs := make([]error, 0)
	if predicted.Min != parsedMin {
		errs = append(errs, fmt.Errorf("min value for city: %s doesn't match", predicted.City))
	}
	if predicted.Avg != parsedAvg {
		errs = append(errs, fmt.Errorf("min value for city: %s doesn't match", predicted.City))
	}
	if predicted.Max != parsedMax {
		errs = append(errs, fmt.Errorf("min value for city: %s doesn't match", predicted.City))
	}
	return errs
}

func collectErrs(errs []error) string {
	result := strings.Builder{}
	for _, err := range errs {
		utils.PanicOnError(result.WriteString(err.Error() + "\n"))
	}
	return result.String()
}
