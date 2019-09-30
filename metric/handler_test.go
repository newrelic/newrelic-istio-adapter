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
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	policy "istio.io/api/policy/v1beta1"
	adapterTest "istio.io/istio/mixer/pkg/adapter/test"
	"istio.io/istio/mixer/template/metric"

	"github.com/newrelic/newrelic-istio-adapter/config"
)

func TestHandleMetric(t *testing.T) {
	testCases := []struct {
		isValid    bool
		metricName string
		value      *policy.Value
	}{
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}},
		},
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}},
		},
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}},
		},
		{
			true,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(-2)}},
		},
		{
			false,
			"not.found.metric",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(0)}},
		},
		{
			false,
			"gaugeExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_EmailAddressValue{EmailAddressValue: &policy.EmailAddress{Value: "yodawg@what.are.you.doing.com"}}},
		},
		{
			true,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}},
		},
		{
			true,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}},
		},
		{
			true,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}},
		},
		{
			false,
			"countExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(-2)}},
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}},
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}},
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}},
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(-123.456)}},
		},
		{
			false,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_StringValue{StringValue: "123ms"}},
		},
		{
			true,
			"summaryExample.instance.istio-system",
			&policy.Value{Value: &policy.Value_DurationValue{DurationValue: &policy.Duration{Value: types.DurationProto(time.Second)}}},
		},
		{
			false,
			"invalid",
			&policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(0)}},
		},
	}

	harvester, err := telemetry.NewHarvester(telemetry.ConfigAPIKey("api-key"), telemetry.ConfigHarvestPeriod(0))
	if err != nil {
		t.Fatal(err)
	}

	metricHandler := &Handler{
		logger: adapterTest.NewEnv(t).Logger(),
		agg:    harvester.MetricAggregator(),
		metrics: map[string]info{
			"gaugeExample.instance.istio-system": info{
				name:  "gaugeExample",
				mtype: config.GAUGE,
			},
			"countExample.instance.istio-system": info{
				name:  "countExample",
				mtype: config.COUNT,
			},
			"summaryExample.instance.istio-system": info{
				name:  "summaryExample",
				mtype: config.SUMMARY,
			},
			"invalid": info{
				name:  "invalid example that shouldn't make it through",
				mtype: config.UNSPECIFIED,
			},
		},
	}

	for _, tc := range testCases {
		i := []*metric.InstanceMsg{
			{
				Name:       tc.metricName,
				Value:      tc.value,
				Dimensions: make(map[string]*policy.Value, 0),
			},
		}
		if err := metricHandler.HandleMetric(context.Background(), i); err != nil {
			if tc.isValid {
				t.Errorf("HandleMetric(%v) errored for valid input: %v", i, err)
			}
		}
	}
}
