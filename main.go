package main

import (
	"fmt"
	"graphql_cache/handlers"
	"graphql_cache/utils/benchmark_utils"
	"graphql_cache/utils/file_utils"
	"reflect"
	"time"
)

const NUMBER_OF_REQUESTS = 3

func main() {
	// call ProxyToCachedAPI three times and log the average time taken to get the result from the function
	// then call the ProxyToAPI three times and log the average time taken to get the result from the function
	// also compare the response bodies for both function calls, and log an error is responses are not the same

	var totalCachedTime, totalAPITime time.Duration
	var cachedResponses, apiResponses []string

	for i := 0; i < NUMBER_OF_REQUESTS; i++ {
		response, elapsed := benchmark_utils.MeasureExecutionTime(handlers.ProxyToAPI)
		fmt.Println(fmt.Sprintf("time taken by ProxyToAPI in attemp %d = %v", i, elapsed))
		totalAPITime += elapsed
		apiResponses = append(apiResponses, string(response))

		response, elapsed = benchmark_utils.MeasureExecutionTime(handlers.ProxyToCachedAPI)
		fmt.Println(fmt.Sprintf("time taken by ProxyToCachedAPI in attemp %d = %v", i, elapsed))
		totalCachedTime += elapsed
		cachedResponses = append(cachedResponses, string(response))
	}

	avgCachedTime := totalCachedTime / NUMBER_OF_REQUESTS
	avgAPITime := totalAPITime / NUMBER_OF_REQUESTS

	fmt.Printf("⏳ Average time for Cached API: %s\n", avgCachedTime)
	fmt.Printf("⏳ Average time for Default API: %s\n", avgAPITime)

	responsesMatch := true
	// Compare responses
	for i := range cachedResponses {
		if reflect.DeepEqual(cachedResponses[i], apiResponses[i]) {
			responsesMatch = false
			break
		}
	}

	if !responsesMatch {
		fmt.Println("❌ Responses are not the same")
	} else {
		fmt.Println("✅ All responses are the same")
	}

	fmt.Println("📝 Generating log files...")
	cachedLog := file_utils.NewFile("output-cached.log.json")
	defaultLog := file_utils.NewFile("output-default.log.json")

	defer cachedLog.Close()
	defer defaultLog.Close()

	cachedLog.WriteJSON(cachedResponses)

	defaultLog.WriteJSON(cachedResponses)

	fmt.Println("✅ Written to log files")
}
