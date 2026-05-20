// Package validator will be used to check if I'm having the correct outputs
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

func ValidateCorrectness(measurements map[string]*model.Measurement) {
	var (
		validation     map[string]interface{}
		totalMinMisses int
		totalMaxMisses int
		totalAvgMisses int
		citiesPassed   int
		citiesFailed   int
	)
	content := utils.PanicE(os.ReadFile("./validation.json"))

	utils.PanicE(struct{}{}, json.Unmarshal(content, &validation))

	for city, temps := range validation { // NOTE: don't think this is right?
		parsedMin, parsedAvg, parsedMax := convertTemperatures(temps.(string))

		predicted, exists := measurements[string(city)]
		if !exists {
			continue
		}
		// utils.PanicOnCondition(!exists, fmt.Sprintf("no data for city: %s", city))

		errs, minMiss, maxMiss, avgMiss := validateNumbers(predicted, parsedMin, parsedAvg, parsedMax)
		totalMinMisses += minMiss
		totalMaxMisses += maxMiss
		totalAvgMisses += avgMiss

		// utils.PanicOnCondition(len(errs) > 0, collectErrs(errs))
		if len(errs) == 0 {
			citiesPassed += 1
			continue
		}
		citiesFailed += 1
		fmt.Println(collectErrs(errs))
	}

	totalMisses := totalMinMisses + totalMaxMisses + totalAvgMisses

	fmt.Println("Finished validating the answers")
	fmt.Printf("Total Misses: %d, Min misses: %d, Max misses: %d, Avg misses: %d\n", totalMisses, totalMinMisses, totalMaxMisses, totalAvgMisses)
	fmt.Printf("Cities Processed: %d, Cities Passed: %d, Cities Failed: %d\n", len(measurements), citiesPassed, citiesFailed)
}

func convertTemperatures(temps string) (float64, float64, float64) {
	values := strings.Split(temps, "/")
	minActual := utils.PanicE(strconv.ParseFloat(values[0], 32))
	avgActual := utils.PanicE(strconv.ParseFloat(values[1], 32))
	maxActual := utils.PanicE(strconv.ParseFloat(values[2], 32))

	parsedMin := utils.TruncateNaive(minActual, 0.1)
	parsedAvg := utils.TruncateNaive(avgActual, 0.1)
	parsedMax := utils.TruncateNaive(maxActual, 0.1)
	return parsedMin, parsedAvg, parsedMax
}

func validateNumbers(predicted *model.Measurement, parsedMin, parsedAvg, parsedMax float64) ([]error, int, int, int) {
	errs := make([]error, 0)
	minMiss, maxMiss, avgMiss := 0, 0, 0
	if predicted.Min != parsedMin {
		minMiss += 1
		errs = append(errs, fmt.Errorf("predicted Min = %.1f, actual = %.1f, city = %v", predicted.Min, parsedMin, predicted.City))
	}
	if predicted.Avg != parsedAvg {
		maxMiss += 1
		errs = append(errs, fmt.Errorf("predicted Avg = %.1f, actual = %.1f, city = %v", predicted.Avg, parsedAvg, predicted.City))
	}
	if predicted.Max != parsedMax {
		avgMiss += 1
		errs = append(errs, fmt.Errorf("predicted Max = %.1f, actual = %.1f, city = %v", predicted.Max, parsedMax, predicted.City))
	}
	return errs, minMiss, maxMiss, avgMiss
}

func collectErrs(errs []error) string {
	result := strings.Builder{}
	for _, err := range errs {
		utils.PanicE(result.WriteString(err.Error() + "\n"))
	}
	return result.String()
}
