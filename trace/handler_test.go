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

package trace

import (
	"reflect"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	policy "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/template/tracespan"
)

func TestConvertTraceSpan(t *testing.T) {
	startTime := time.Date(2020, time.August, 10, 11, 12, 30, 0, time.UTC)
	startTimestamp, err := types.TimestampProto(startTime)
	if err != nil {
		t.Errorf("error creating startTimestamp: %v", err)
	}

	endTime := time.Date(2020, time.August, 10, 11, 12, 31, 0, time.UTC)
	endTimestamp, err := types.TimestampProto(endTime)
	if err != nil {
		t.Errorf("error creating endTimestamp: %v", err)
	}

	spanTags := map[string]*policy.Value{
		"zip":   &policy.Value{Value: &policy.Value_StringValue{StringValue: "zap"}},
		"int":   &policy.Value{Value: &policy.Value_Int64Value{Int64Value: 123}},
		"float": &policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: 4.56}},
	}

	tSpan := &tracespan.InstanceMsg{
		SpanId:              "some-guid-value",
		TraceId:             "123",
		Name:                "name",
		ParentSpanId:        "456",
		StartTime:           &policy.TimeStamp{Value: startTimestamp},
		EndTime:             &policy.TimeStamp{Value: endTimestamp},
		SpanName:            "span-name",
		HttpStatusCode:      200,
		ClientSpan:          true,
		RewriteClientSpanId: true,
		SourceName:          "source-service",
		SourceIp:            &policy.IPAddress{Value: []byte{192, 168, 2, 2}},
		DestinationName:     "destination-service",
		DestinationIp:       &policy.IPAddress{Value: []byte{192, 168, 3, 3}},
		RequestSize:         int64(99),
		RequestTotalSize:    int64(199),
		ResponseSize:        int64(299),
		ResponseTotalSize:   int64(399),
		ApiProtocol:         "http",
		SpanTags:            spanTags,
	}

	expectedAttrs := map[string]interface{}{
		"response.code":       int64(200),
		"clientSpan":          true,
		"rewriteClientSpanId": true,
		"source.name":         "source-service",
		"source.ip":           "192.168.2.2",
		"destination.name":    "destination-service",
		"destination.ip":      "192.168.3.3",
		"request.size":        int64(99),
		"request.totalSize":   int64(199),
		"response.size":       int64(299),
		"response.totalSize":  int64(399),
		"api.protocol":        "http",
		"zip":                 "zap",
		"int":                 int64(123),
		"float":               float64(4.56),
	}

	expected := &telemetry.Span{
		ID:          "some-guid-value",
		TraceID:     "123",
		Name:        "span-name",
		ParentID:    "456",
		Timestamp:   startTime,
		Duration:    time.Second,
		ServiceName: "source-service",
		Attributes:  expectedAttrs,
	}

	actual, err := convertTraceSpan(tSpan)
	if err != nil {
		t.Errorf("failed to convert tracespan: %v", err)
	}

	if actual.ID != expected.ID {
		t.Errorf("expected GUID '%s', got '%s'", expected.ID, actual.ID)
	}

	if actual.TraceID != expected.TraceID {
		t.Errorf("expected TraceID '%s', got '%s'", expected.TraceID, actual.TraceID)
	}

	if actual.Name != expected.Name {
		t.Errorf("expected Name '%s', got '%s'", expected.Name, actual.Name)
	}

	if actual.ParentID != expected.ParentID {
		t.Errorf("expected ParentID '%s', got '%s'", expected.ParentID, actual.ParentID)
	}

	if actual.Timestamp != expected.Timestamp {
		t.Errorf("expected Timestamp '%s', got '%s'", expected.Timestamp, actual.Timestamp)
	}

	if actual.Duration != expected.Duration {
		t.Errorf("expected Duration '%s', got '%s'", expected.Duration, actual.Duration)
	}

	if actual.ServiceName != expected.ServiceName {
		t.Errorf("expected ServiceName '%s', got '%s'", expected.ServiceName, actual.ServiceName)
	}

	if !reflect.DeepEqual(actual.Attributes, expected.Attributes) {
		t.Errorf("expected attributes '%#v', got '%#v'", expected.Attributes, actual.Attributes)
	}
}
