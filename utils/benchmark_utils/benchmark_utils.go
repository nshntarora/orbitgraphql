package benchmark_utils

import "time"

func MeasureExecutionTime(fn func() []byte) ([]byte, time.Duration) {
	start := time.Now()
	response := fn()
	elapsed := time.Since(start)
	return response, elapsed
}
