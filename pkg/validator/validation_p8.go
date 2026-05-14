// Package validator will be used to check if I'm having the correct outputs
package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/throwea/1brc-go/pkg/model"
	"github.com/throwea/1brc-go/pkg/utils"
)

func ValidateCorrectnessInt(measurements map[string]*model.Predicted) {
	getActual := func(temps, city string) *model.Actual {
		split := strings.Split(temps, "/")
		min, avg, max := 0, 1, 2
		return &model.Actual{
			City: city,
			Min:  split[min],
			Max:  split[max],
			Avg:  split[avg],
		}
	}
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
		actual := getActual(temps.(string), city)

		predicted, exists := measurements[city]
		if !exists {
			panic(city + " data not found")
		}

		errs, minMiss, maxMiss, avgMiss := compare(predicted, actual)
		totalMinMisses += minMiss
		totalMaxMisses += maxMiss
		totalAvgMisses += avgMiss

		// utils.PanicOnCondition(len(errs) > 0, collectErrs(errs))
		if len(errs) == 0 {
			citiesPassed += 1
			continue
		}
		citiesFailed += 1
		fmt.Println(Errors(errs))
	}

	totalMisses := totalMinMisses + totalMaxMisses + totalAvgMisses

	fmt.Println("Finished validating the answers")
	fmt.Printf("Total Misses: %d, Min misses: %d, Max misses: %d, Avg misses: %d\n", totalMisses, totalMinMisses, totalMaxMisses, totalAvgMisses)
	fmt.Printf("Cities Processed: %d, Cities Passed: %d, Cities Failed: %d\n", len(measurements), citiesPassed, citiesFailed)
}

func compare(predicted *model.Predicted, actual *model.Actual) ([]error, int, int, int) {
	errs := make([]error, 0)
	minMiss, maxMiss, avgMiss := 0, 0, 0
	if predicted.Min != actual.Min {
		minMiss += 1
		errs = append(errs, fmt.Errorf("predicted Min = %s, actual = %s, city = %v", predicted.Min, actual.Min, predicted.City))
	}
	if predicted.Avg != actual.Avg {
		maxMiss += 1
		errs = append(errs, fmt.Errorf("predicted Avg = %s, actual = %s, city = %v", predicted.Avg, actual.Avg, predicted.City))
	}
	if predicted.Max != actual.Max {
		avgMiss += 1
		errs = append(errs, fmt.Errorf("predicted Max = %s, actual = %s, city = %v", predicted.Max, actual.Max, predicted.City))
	}
	return errs, minMiss, maxMiss, avgMiss
}

func Errors(errs []error) string {
	result := strings.Builder{}
	for _, err := range errs {
		utils.PanicE(result.WriteString(err.Error() + "\n"))
	}
	return result.String()
}
