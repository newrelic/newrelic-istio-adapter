package instrumentation

import (
	"testing"
	"time"
)

func TestSummary(t *testing.T) {
	ag := NewMetricAggregator()
	summary := ag.NewSummary("mySummary", map[string]interface{}{"zip": "zap"})
	summary.Record(3)
	summary.Record(4)
	summary.Record(5)

	expect := `[{"Name":"mySummary","Attributes":null,"AttributesJSON":{"zip":"zap"},"Count":3,"Sum":12,"Min":3,"Max":5,"Timestamp":"0001-01-01T00:00:00Z","Interval":0}]`
	testMetrics(t, ag, expect)
}

func TestSummaryDuration(t *testing.T) {
	ag := NewMetricAggregator()
	summary := ag.NewSummary("mySummary", map[string]interface{}{"zip": "zap"})
	summary.RecordDuration(3 * time.Second)
	summary.RecordDuration(4 * time.Second)
	summary.RecordDuration(5 * time.Second)

	expect := `[{"Name":"mySummary","Attributes":null,"AttributesJSON":{"zip":"zap"},"Count":3,"Sum":12000,"Min":3000,"Max":5000,"Timestamp":"0001-01-01T00:00:00Z","Interval":0}]`
	testMetrics(t, ag, expect)
}

func TestNilAggregatorSummaries(t *testing.T) {
	var ag *MetricAggregator
	summary := ag.NewSummary("summary", map[string]interface{}{})
	summary.Record(1)
	summary.RecordDuration(time.Second)
}

func TestNilSummaryMethods(t *testing.T) {
	var summary *Summary
	summary.Record(1)
	summary.RecordDuration(time.Second)
}

func TestSummaryNilAggregator(t *testing.T) {
	s := Summary{}
	s.Record(10)
	s.RecordDuration(time.Second)
}

func TestSummaryMinMax(t *testing.T) {
	ag := NewMetricAggregator()
	s := ag.NewSummary("sum", nil)
	s.Record(2)
	s.Record(1)
	s.Record(3)
	expect := `[{"Name":"sum","Attributes":null,"AttributesJSON":{},"Count":3,"Sum":6,"Min":1,"Max":3,"Timestamp":"0001-01-01T00:00:00Z","Interval":0}]`
	testMetrics(t, ag, expect)
}
