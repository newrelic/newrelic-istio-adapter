// Copyright 2019 New Relic Corporation
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
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/instrumentation"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
	policy "istio.io/api/policy/v1beta1"
	adapterTest "istio.io/istio/mixer/pkg/adapter/test"
	"istio.io/istio/mixer/template/metric"

	"github.com/newrelic/newrelic-istio-adapter/config"
)

func TestHandleMetric(t *testing.T) {
	testCases := []struct {
		isValid       bool
		metricName    string
		value         *policy.Value
		expectedName  string
		expectedValue float64
	}{
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}},
			"gaugeExample",
			float64(1.0),
		},
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}},
			"gaugeExample",
			float64(2.0),
		},
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}},
			"gaugeExample",
			float64(3.33),
		},
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(-2)}},
			"gaugeExample",
			float64(-2.0),
		},
		{
			true,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}},
			"countExample",
			float64(1.0),
		},
		{
			true,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}},
			"countExample",
			float64(2.0),
		},
		{
			true,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}},
			"countExample",
			float64(3.33),
		},
		{
			false,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(-2)}},
			"countExample",
			float64(0),
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}},
			"summaryExample",
			float64(1.0),
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}},
			"summaryExample",
			float64(2.0),
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}},
			"summaryExample",
			float64(3.33),
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(-123.456)}},
			"summaryExample",
			float64(-123.456),
		},
		{
			false,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "123ms"}},
			"summaryExample",
			float64(123.0),
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DurationValue{DurationValue: &policy.Duration{Value: types.DurationProto(time.Second)}}},
			"summaryExample",
			float64(1000.0),
		},
	}

	handlerMetrics := make(map[string]info, 3)
	handlerMetrics["gaugeExample.instance.istio-system"] = info{
		name:  "gaugeExample",
		mtype: config.GAUGE,
	}
	handlerMetrics["countExample.instance.istio-system"] = info{
		name:  "countExample",
		mtype: config.COUNT,
	}
	handlerMetrics["summaryExample.instance.istio-system"] = info{
		name:  "summaryExample",
		mtype: config.SUMMARY,
	}

	metricHandler := &Handler{
		logger:  adapterTest.NewEnv(t).Logger(),
		agg:     instrumentation.NewMetricAggregator(),
		metrics: handlerMetrics,
	}

	for _, tc := range testCases {
		instances := make([]*metric.InstanceMsg, 1)
		instances[0] = &metric.InstanceMsg{
			Name:       tc.metricName,
			Value:      tc.value,
			Dimensions: make(map[string]*policy.Value, 0),
		}
		err := metricHandler.HandleMetric(context.Background(), instances)
		if err != nil {
			t.Errorf("HandleMetric() error: %v", err)
		}

		metrics := metricHandler.agg.Metrics()
		if len(metrics) == 0 && !tc.isValid {
			// We correctly skipped invalid metric, so nothing left to test.
			continue
		}
		if len(metrics) == 0 && tc.isValid {
			t.Errorf("failed to record valid metric %q with value %v.", tc.metricName, tc.value)
			continue
		}
		if len(metrics) > 1 {
			t.Errorf("expected to record only one metric %q, got %v", tc.metricName, metrics)
			continue
		}
		metric := metrics[0]

		switch m := metric.(type) {
		case *telemetry.Gauge:
			if m.Name != tc.expectedName {
				t.Errorf("expected name %s, got %s", tc.expectedName, m.Name)
			}
			if m.Value != tc.expectedValue {
				t.Errorf("expected value %f, got %f", tc.expectedValue, m.Value)
			}
		case *telemetry.Count:
			if m.Name != tc.expectedName {
				t.Errorf("expected name %s, got %s", tc.expectedName, m.Name)
			}
			if m.Value != tc.expectedValue {
				t.Errorf("expected value %f, got %f", tc.expectedValue, m.Value)
			}
		case *telemetry.Summary:
			if m.Name != tc.expectedName {
				t.Errorf("expected name %s, got %s", tc.expectedName, m.Name)
			}
			if m.Sum != tc.expectedValue {
				t.Errorf("expected value %f, got %f", tc.expectedValue, m.Sum)
			}
		default:
			t.Errorf("unknown metric type %T for %s", m, tc.metricName)
		}

	}
}
