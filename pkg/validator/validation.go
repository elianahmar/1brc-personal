package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/throwea/1brc-go/pkg/model"
)

func validateCorrectness(measurements map[city]model.measurement) {
	var validation map[city]model.Measurement
	content, err := os.ReadFile("./validation.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(content, &validation); err != nil {
		panic(err)
	}
	for city, temps := range measurements {
		parsedMin, parsedAvg, parsedMax := convertTemperatures(temps)
		predicted, exists := measurements[city]
		if !exists {
			panic(fmt.Errorf("no data for city: %s", city))
		}
		errs := validateNumbers(predicted, parsedMin, parsedAvg, parsedMax)
		if len(errs) > 0 {
			panic(collectErrs(errs))
		}

	}
}

func convertTemperatures(temps string) (float64, float64, float64) {
	values := strings.Split(temps, "/")
	minActual, _ := strconv.ParseFloat(values[0], 32)
	avgActual, _ := strconv.ParseFloat(values[1], 32)
	maxActual, _ := strconv.ParseFloat(values[2], 32)
	parsedMin := truncateNaive(minActual, 0.1)
	parsedAvg := truncateNaive(avgActual, 0.1)
	parsedMax := truncateNaive(maxActual, 0.1)
	return parsedMin, parsedAvg, parsedMax
}

func validateNumbers(predicted Validation, parsedMin, parsedAvg, parsedMax float64) []error {
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
	return nil
}

func collectErrs(errs []error) strings.Builder {
	result := strings.Builder{}
	for _, err := range errs {
		_, err := result.WriteString(err.Error() + "\n")
		if err != nil {
			panic("can't write the errors to a string")
		}
	}
	return result
}
