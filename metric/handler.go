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
	"errors"
	"fmt"

	"github.com/newrelic/newrelic-istio-adapter/config"
	"github.com/newrelic/newrelic-istio-adapter/convert"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"istio.io/istio/mixer/template/metric"
)

// Handler represents a processor that can handle metrics from Istio and transmit them to New Relic.
type Handler struct {
	agg     *telemetry.MetricAggregator
	metrics map[string]info
}

type info struct {
	name  string
	mtype config.Params_MetricInfo_Type
}

type handleError struct {
	Name string
	Err  error
}

func (e *handleError) Error() string {
	return fmt.Sprintf("failed to handle %q: %s", e.Name, e.Err)
}

type handleErrors []*handleError

func (e handleErrors) Error() string {
	if e == nil || len(e) == 0 {
		return ""
	}

	var msg string
	for _, err := range e {
		msg += "\n"
		msg += err.Error()
	}
	return msg
}

func (e handleErrors) ErrorOrNil() error {
	if e == nil || len(e) == 0 {
		return nil
	}
	return e
}

// HandleMetric transforms metric template instances into New Relic metrics and
// sends them to New Relic.
func (h *Handler) HandleMetric(_ context.Context, msgs []*metric.InstanceMsg) error {
	var errs handleErrors
	for _, i := range msgs {
		v, err := convert.ValueToFloat64(i.Value)
		if err != nil {
			errs = append(errs, &handleError{i.Name, err})
			continue
		}
		attrs := convert.DimensionsToAttributes(i.Dimensions)

		minfo, found := h.metrics[i.Name]
		if !found {
			errs = append(errs, &handleError{i.Name, errors.New("no metric info found")})
			continue
		}

		switch minfo.mtype {
		case config.GAUGE:
			h.agg.Gauge(minfo.name, attrs).Value(v)
		case config.COUNT:
			if v < 0.0 {
				errs = append(errs, &handleError{i.Name, fmt.Errorf("negative count value: %f", v)})
				continue
			}
			h.agg.Count(minfo.name, attrs).Increase(v)
		case config.SUMMARY:
			h.agg.Summary(minfo.name, attrs).Record(v)
		default:
			errs = append(errs, &handleError{i.Name, fmt.Errorf("unknown metric type: %v", minfo.mtype)})
		}
	}

	return errs.ErrorOrNil()
}
