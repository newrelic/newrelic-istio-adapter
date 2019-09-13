// Copyright 2019 New Relic Corporation
// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

import (
	"context"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/instrumentation"
	"github.com/newrelic/newrelic-istio-adapter/config"
	"github.com/newrelic/newrelic-istio-adapter/convert"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/metric"
)

// Handler represents a processor that can handle metrics from Istio and transmit them to New Relic.
type Handler struct {
	logger  adapter.Logger
	agg     *instrumentation.MetricAggregator
	metrics map[string]info
}

type info struct {
	name  string
	mtype config.Params_MetricInfo_Type
}

// HandleMetric transforms metric template instances into New Relic metrics and
// sends them to New Relic.
func (h *Handler) HandleMetric(_ context.Context, msgs []*metric.InstanceMsg) error {
	for _, i := range msgs {
		v, err := convert.ValueToFloat64(i.Value)
		if err != nil {
			h.logger.Warningf("Failed to parse metric value '%v' for '%s'", i.Value, i.Name)
			continue
		}
		attrs := convert.DimensionsToAttributes(i.Dimensions)

		minfo, found := h.metrics[i.Name]
		if !found {
			h.logger.Warningf("no metric info found for %s, skipping metric", i.Name)
			continue
		}

		switch minfo.mtype {
		case config.GAUGE:
			h.agg.NewGauge(minfo.name, attrs).Value(v)
		case config.COUNT:
			if v < 0.0 {
				h.logger.Warningf("invalid count value for %s (skipping): must be a positive value for a count (got %f)", i.Name, v)
				continue
			}
			h.agg.NewCount(minfo.name, attrs).Increase(v)
		case config.SUMMARY:
			h.agg.NewSummary(minfo.name, attrs).Record(v)
		default:
			h.logger.Warningf("unknown metric type for %s: %v", i.Name, minfo.mtype)
		}
	}

	return nil
}
