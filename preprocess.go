package main

func collectData()
	measurements := make(map[city]measurement)
	for text := range data {
		measurement := processLine(text)
		split := strings.Split(text, ";")
		city := city(split[0])
		if _, exists := measurements[city]; !exists {
			measurements[city] = measurement{}
		}
		measurements[city].temp += measurement.temp
		measurements[city].count += 1
		fmt.Printf("%v\n", measurement)
	}
}
	// }(data)
