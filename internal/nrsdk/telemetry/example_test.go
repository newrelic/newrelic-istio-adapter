package telemetry

import (
	"encoding/json"
	"os"
	"time"
)

func Example() {
	h := NewHarvester(
		ConfigAPIKey(os.Getenv("NEW_RELIC_API_KEY")),
		ConfigCommonAttributes(map[string]interface{}{
			"app.name": "myApplication",
		}),
		ConfigBasicErrorLogger(os.Stderr),
		ConfigBasicDebugLogger(os.Stdout),
	)

	start := time.Now()
	duration := time.Second

	// raw metric
	h.RecordMetric(Gauge{
		Timestamp: time.Now(),
		Value:     1,
		Name:      "myMetric",
		Attributes: map[string]interface{}{
			"color": "purple",
		},
	})

	// span
	h.RecordSpan(Span{
		ID:          "12345",
		TraceID:     "67890",
		Name:        "purple-span",
		Timestamp:   start,
		Duration:    duration,
		ServiceName: "ExampleApplication",
		Attributes: map[string]interface{}{
			"color": "purple",
		},
	})
}

func ExampleNewHarvester() {
	h := NewHarvester(
		ConfigAPIKey(os.Getenv("NEW_RELIC_API_KEY")),
	)
	h.RecordMetric(Gauge{
		Timestamp: time.Now(),
		Value:     1,
		Name:      "myMetric",
		Attributes: map[string]interface{}{
			"color": "purple",
		},
	})
}

func ExampleHarvester_RecordMetric() {
	h := NewHarvester(
		ConfigAPIKey(os.Getenv("NEW_RELIC_API_KEY")),
	)
	start := time.Now()
	h.RecordMetric(Count{
		Name:           "myCount",
		AttributesJSON: json.RawMessage(`{"zip":"zap"}`),
		Value:          123,
		Timestamp:      start,
		Interval:       5 * time.Second,
	})
}
