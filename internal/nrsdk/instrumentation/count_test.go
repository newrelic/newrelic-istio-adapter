package instrumentation

import (
	"testing"
)

func TestCount(t *testing.T) {
	ag := NewMetricAggregator()
	count := ag.NewCount("myCount", map[string]interface{}{"zip": "zap"})
	count.Increase(22.5)
	count.Increment()

	expect := `[{"Name":"myCount","Attributes":null,"AttributesJSON":{"zip":"zap"},"Value":23.5,"Timestamp":"0001-01-01T00:00:00Z","Interval":0}]`
	testMetrics(t, ag, expect)
}

func TestCountNegative(t *testing.T) {
	ag := NewMetricAggregator()
	count := ag.NewCount("myCount", map[string]interface{}{"zip": "zap"})
	count.Increase(-123)
	if ms := ag.Metrics(); len(ms) != 0 {
		t.Fatal(ms)
	}
}

func TestNilAggregatorCounts(t *testing.T) {
	var ag *MetricAggregator
	count := ag.NewCount("count", map[string]interface{}{})
	count.Increment()
	count.Increase(5)
}

func TestNilCountMethods(t *testing.T) {
	var count *Count
	count.Increment()
	count.Increase(5)
}

func TestCountNilAggregator(t *testing.T) {
	c := Count{}
	c.Increment()
	c.Increase(1)
}
