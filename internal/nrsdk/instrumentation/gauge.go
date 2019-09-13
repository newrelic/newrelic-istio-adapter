package instrumentation

import (
	"encoding/json"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
)

// Gauge is the metric type that records a value that can increase or decrease.
// It generally represents the value for something at a particular moment in
// time.  One typically records a Gauge value on a set interval.
//
// Only the most recent Gauge metric value is reported over a given harvest
// period, all others are dropped.
//
// Example possible uses:
//
//  * the temperature in a room
//  * the amount of memory currently in use for a process
//  * the bytes per second flowing into Kafka at this exact moment in time
//  * the current speed of your car
//
type Gauge struct{ metricHandle }

// valueNow facilitates testing.
func (g *Gauge) valueNow(val float64, now time.Time) {
	if nil == g {
		return
	}
	ag := g.aggregator
	if nil == ag {
		return
	}

	ag.lock.Lock()
	defer ag.lock.Unlock()

	m := ag.findOrCreateMetric(g.metricIdentity)
	if nil == m.g {
		m.g = &telemetry.Gauge{
			Name:           g.Name,
			AttributesJSON: json.RawMessage(g.attributesJSON),
			Value:          val,
		}
	}
	m.g.Value = val
	m.g.Timestamp = now
}

// Value records the value given.
func (g *Gauge) Value(val float64) {
	g.valueNow(val, time.Now())
}
