package instrumentation

import (
	"encoding/json"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
)

// Summary is the metric type used for reporting aggregated information about
// discrete events.   It provides the count, average, sum, min and max values
// over time.  All fields are reset to 0 every reporting interval.
//
// The final metric reported at the end of a harvest period is an aggregation.
// Values reported are the count of the number of metrics recorded, sum of
// all their values, minimum value recorded, and maximum value recorded.
//
// Example possible uses:
//
//  * the duration and count of spans
//  * the duration and count of transactions
//  * the time each message spent in a queue
//
type Summary struct{ metricHandle }

// Record adds an observation to a summary.
func (s *Summary) Record(val float64) {
	if nil == s {
		return
	}
	ag := s.aggregator
	if nil == ag {
		return
	}

	ag.lock.Lock()
	defer ag.lock.Unlock()

	m := ag.findOrCreateMetric(s.metricIdentity)
	if nil == m.s {
		m.s = &telemetry.Summary{
			Name:           s.Name,
			AttributesJSON: json.RawMessage(s.attributesJSON),
			Count:          1,
			Sum:            val,
			Min:            val,
			Max:            val,
		}
		return
	}
	m.s.Sum += val
	m.s.Count++
	if val < m.s.Min {
		m.s.Min = val
	}
	if val > m.s.Max {
		m.s.Max = val
	}
}

// RecordDuration adds a duration observation to a summary.  It records the
// value in milliseconds, New Relic's recommended duration units.
func (s *Summary) RecordDuration(val time.Duration) {
	s.Record(val.Seconds() * 1000.0)
}
