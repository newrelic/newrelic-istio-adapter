package instrumentation

import (
	"encoding/json"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
)

// Count is the metric type that counts the number of times an event occurred.
// This counter is reset every time the data is reported, meaning the value
// reported represents the difference in count over the reporting time window.
//
// Example possible uses:
//
//  * the number of messages put on a topic
//  * the number of HTTP requests
//  * the number of errors thrown
//  * the number of support tickets answered
//
type Count struct{ metricHandle }

// Increment increases the Count value by one.
func (c *Count) Increment() {
	c.Increase(1)
}

// Increase increases the Count value by the number given.  The value must be
// non-negative.
func (c *Count) Increase(val float64) {
	if nil == c {
		return
	}
	ag := c.aggregator
	if nil == ag {
		return
	}

	if val < 0 {
		return
	}

	ag.lock.Lock()
	defer ag.lock.Unlock()

	m := ag.findOrCreateMetric(c.metricIdentity)
	if nil == m.c {
		m.c = &telemetry.Count{
			Name:           c.Name,
			AttributesJSON: json.RawMessage(c.attributesJSON),
		}
	}
	m.c.Value += val
}
