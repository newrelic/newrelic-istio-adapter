package instrumentation

import (
	"os"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
)

func Example() {
	// Create a MetricAggregator:
	agg := NewMetricAggregator()
	// Then create a telemetry.NewHarvester.  Be sure to register the
	// aggregator with the harvester by using the aggregator's
	// BeforeHarvest method.
	telemetry.NewHarvester(
		telemetry.ConfigAPIKey(os.Getenv("NEW_RELIC_API_KEY")),
		agg.BeforeHarvest,
	)

	// Now use the aggregator to create metrics.  You can create metrics in
	// a single line:
	agg.NewGauge("temperature", map[string]interface{}{"zip": "zap"}).Value(23.4)

	// Or use a reference to the metric's identity to add data-points with
	// to the same metric with minimal overhead:
	count := agg.NewCount("iterations", map[string]interface{}{"zip": "zap"})
	for {
		count.Increment()
	}
	// When the harvester harvests (by default every 5 seconds), it will
	// collect metrics from the MetricAggregator and send them to New Relic.
}

func ExampleMetricAggregator_NewCount() {
	ag := NewMetricAggregator()
	count := ag.NewCount("myCount", map[string]interface{}{"zip": "zap"})
	count.Increment()
}

func ExampleMetricAggregator_NewGauge() {
	ag := NewMetricAggregator()
	gauge := ag.NewGauge("temperature", map[string]interface{}{"zip": "zap"})
	gauge.Value(23.4)
}

func ExampleMetricAggregator_NewSummary() {
	ag := NewMetricAggregator()
	summary := ag.NewSummary("mySummary", map[string]interface{}{"zip": "zap"})
	summary.RecordDuration(3 * time.Second)
}

func ExampleMetricAggregator_BeforeHarvest() {
	// Create a MetricAggregator
	agg := NewMetricAggregator()
	// Register this aggregator with a Harvester. This is necessary if you want
	// to harvest metrics on a schedule.
	telemetry.NewHarvester(
		telemetry.ConfigAPIKey(os.Getenv("NEW_RELIC_API_KEY")),
		agg.BeforeHarvest,
	)
}
