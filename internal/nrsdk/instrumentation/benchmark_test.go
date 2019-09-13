package instrumentation

import "testing"

func BenchmarkAddMetric(b *testing.B) {
	// This benchmark tests creating and aggregating a summary.
	ag := NewMetricAggregator()
	attributes := map[string]interface{}{"zip": "zap", "zop": 123}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		summary := ag.NewSummary("mySummary", attributes)
		summary.Record(12.3)
		if nil == summary {
			b.Fatal("nil summary")
		}
	}
}
