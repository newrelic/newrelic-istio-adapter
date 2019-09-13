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

package convert

import (
	"net"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	policy "istio.io/api/policy/v1beta1"
)

func TestValueToFloat64(t *testing.T) {
	testCases := []struct {
		isValid  bool
		value    *policy.Value
		expected float64
	}{
		{true, &policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}}, float64(1.0)},
		{false, &policy.Value{Value: &policy.Value_StringValue{StringValue: "love"}}, float64(0.0)},
		{true, &policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}}, float64(2.0)},
		{true, &policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}}, float64(3.33)},
		{true, &policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(-2)}}, float64(-2.0)},
		{true, &policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(-123.456)}}, float64(-123.456)},
		{true, &policy.Value{Value: &policy.Value_DurationValue{DurationValue: &policy.Duration{
			Value: types.DurationProto(time.Second),
		}}}, float64(1000.0)},
		{false, &policy.Value{Value: &policy.Value_BoolValue{BoolValue: true}}, float64(0.0)},
	}

	for _, tc := range testCases {
		result, err := ValueToFloat64(tc.value)
		if err != nil && tc.isValid {
			t.Errorf("decodeMetricValue() error: %v", err)
		}

		if result != tc.expected && tc.isValid {
			t.Errorf("failed to decode valid value %v: expected %v got %v.", tc.value, tc.expected, result)
			continue
		}
		if result != 0.0 && !tc.isValid {
			t.Errorf("decoded an invalid value: expected 0.0 got value %v.", result)
		}
	}
}

func TestValueToAttribute(t *testing.T) {
	testTime, err := types.TimestampProto(time.Unix(1582230020, 0))
	if err != nil {
		t.Fatalf("failed to parse timestamp test data: %v", err)
	}

	testCases := []struct {
		isValid  bool
		value    *policy.Value
		expected interface{}
	}{
		{true, &policy.Value{Value: &policy.Value_StringValue{StringValue: "1"}}, "1"},
		{true, &policy.Value{Value: &policy.Value_StringValue{StringValue: "love"}}, "love"},
		{true, &policy.Value{Value: &policy.Value_Int64Value{Int64Value: int64(2)}}, int64(2)},
		{true, &policy.Value{Value: &policy.Value_DoubleValue{DoubleValue: float64(3.33)}}, float64(3.33)},
		{true, &policy.Value{Value: &policy.Value_DurationValue{DurationValue: &policy.Duration{
			Value: types.DurationProto(time.Second),
		}}}, float64(1000.0)},
		{true, &policy.Value{Value: &policy.Value_BoolValue{BoolValue: true}}, true},
		{true, &policy.Value{Value: &policy.Value_TimestampValue{TimestampValue: &policy.TimeStamp{
			Value: testTime,
		}}}, float64(1582230020000)},
		{true, &policy.Value{Value: &policy.Value_IpAddressValue{IpAddressValue: &policy.IPAddress{
			Value: net.ParseIP("127.0.0.1"),
		}}}, "127.0.0.1"},
		{true, &policy.Value{Value: &policy.Value_EmailAddressValue{EmailAddressValue: &policy.EmailAddress{
			Value: "new@rel.ic",
		}}}, "new@rel.ic"},
		{true, &policy.Value{Value: &policy.Value_DnsNameValue{DnsNameValue: &policy.DNSName{
			Value: "newrelic.ninja",
		}}}, "newrelic.ninja"},
		{true, &policy.Value{Value: &policy.Value_UriValue{UriValue: &policy.Uri{
			Value: "file:///etc/shadow",
		}}}, "file:///etc/shadow"},
		{true, &policy.Value{Value: &policy.Value_StringMapValue{StringMapValue: &policy.StringMap{
			Value: map[string]string{"hello": "aloha", "goodbye": "aloha"},
		}}}, "goodbye=aloha&hello=aloha"},
	}

	for _, tc := range testCases {
		result := ValueToAttribute(tc.value)

		if result != tc.expected && tc.isValid {
			t.Errorf("failed to decode valid value %v: expected %v (%T) got %v (%T).",
				tc.value, tc.expected, tc.expected, result, result)
			continue
		}
		if result == tc.expected && !tc.isValid {
			t.Errorf("decoded an unknown attribute type %v {%T) : got %v (%T)",
				tc.value, tc.value, result, result)
		}
	}
}
