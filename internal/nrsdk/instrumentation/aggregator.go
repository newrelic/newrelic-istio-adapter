package instrumentation

import (
	"sync"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/internal"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
)

type metricIdentity struct {
	// Note that the type is not a field here since a single 'metric' type
	// may contain a count, gauge, and summary.
	Name           string
	attributesJSON string
}

type metric struct {
	s *telemetry.Summary
	c *telemetry.Count
	g *telemetry.Gauge
}

type metricHandle struct {
	metricIdentity
	aggregator *MetricAggregator
}

func newMetricHandle(ag *MetricAggregator, name string, attributes map[string]interface{}) metricHandle {
	return metricHandle{
		aggregator: ag,
		metricIdentity: metricIdentity{
			attributesJSON: string(internal.MarshalOrderedAttributes(attributes)),
			Name:           name,
		},
	}
}

// MetricAggregator combines individual datapoints into metrics.
type MetricAggregator struct {
	lock    sync.Mutex
	metrics map[metricIdentity]*metric
}

// NewMetricAggregator creates a new MetricAggregator.
func NewMetricAggregator() *MetricAggregator {
	return &MetricAggregator{
		metrics: make(map[metricIdentity]*metric),
	}
}

// findOrCreateMetric finds or creates the metric associated with the given
// identity.  This function assumes the Harvester is locked.
func (ag *MetricAggregator) findOrCreateMetric(identity metricIdentity) *metric {
	m := ag.metrics[identity]
	if nil == m {
		// this happens the first time we update the value,
		// or after a harvest when the metric is removed.
		m = &metric{}
		ag.metrics[identity] = m
	}
	return m
}

// NewCount creates a new Count metric.
func (ag *MetricAggregator) NewCount(name string, attributes map[string]interface{}) *Count {
	return &Count{metricHandle: newMetricHandle(ag, name, attributes)}
}

// NewGauge creates a new Gauge metric.
func (ag *MetricAggregator) NewGauge(name string, attributes map[string]interface{}) *Gauge {
	return &Gauge{metricHandle: newMetricHandle(ag, name, attributes)}
}

// NewSummary creates a new Summary metric.
func (ag *MetricAggregator) NewSummary(name string, attributes map[string]interface{}) *Summary {
	return &Summary{metricHandle: newMetricHandle(ag, name, attributes)}
}

// Metrics returns the metrics that have been added to the aggregator since the
// last call to Metrics. Once those metrics are returned, the aggregator is
// reset and metric aggregation will begin anew.
func (ag *MetricAggregator) Metrics() []telemetry.Metric {
	ag.lock.Lock()
	mts := ag.metrics
	ag.metrics = make(map[metricIdentity]*metric, len(mts))
	ag.lock.Unlock()

	var ms []telemetry.Metric
	for _, m := range mts {
		if nil != m.c {
			ms = append(ms, m.c)
		}
		if nil != m.s {
			ms = append(ms, m.s)
		}
		if nil != m.g {
			ms = append(ms, m.g)
		}
	}
	return ms
}

// BeforeHarvest registers the aggregator with the harvester so that the
// aggregated metrics will be harvested on the harvester's interval. This is
// the preferred way to send metrics from a MetricAggregator to New Relic.
func (ag *MetricAggregator) BeforeHarvest(cfg *telemetry.Config) {
	previous := cfg.BeforeHarvestFunc
	cfg.BeforeHarvestFunc = func(h *telemetry.Harvester) {
		if nil != previous {
			previous(h)
		}
		for _, m := range ag.Metrics() {
			h.RecordMetric(m)
		}
	}
}
