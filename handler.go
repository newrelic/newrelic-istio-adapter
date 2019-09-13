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

package newrelic

import (
	"context"

	nrmetric "github.com/newrelic/newrelic-istio-adapter/metric"
	"github.com/newrelic/newrelic-istio-adapter/trace"
	"istio.io/istio/mixer/template/metric"
	"istio.io/istio/mixer/template/tracespan"
)

// Handler represents a processor that can handle instances from Istio and transmit them to New Relic.
type Handler struct {
	m *nrmetric.Handler
	t *trace.Handler
}

// HandleMetric implements the HandleMetricServiceServer interface.
func (h *Handler) HandleMetric(ctx context.Context, values []*metric.InstanceMsg) error {
	return h.m.HandleMetric(ctx, values)
}

// HandleTraceSpan implements the HandleTraceSpanServiceServer interface.
func (h *Handler) HandleTraceSpan(ctx context.Context, values []*tracespan.InstanceMsg) error {
	return h.t.HandleTraceSpan(ctx, values)
}
