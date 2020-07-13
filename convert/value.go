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
	"fmt"
	"strconv"

	policy "istio.io/api/policy/v1beta1"
	"istio.io/istio/mixer/pkg/adapter"
)

// ValueToFloat64 returns a metric value as a float64.
func ValueToFloat64(in *policy.Value) (float64, error) {
	switch t := in.GetValue().(type) {
	case *policy.Value_StringValue:
		return strconv.ParseFloat(t.StringValue, 64)
	case *policy.Value_Int64Value:
		return float64(t.Int64Value), nil
	case *policy.Value_DoubleValue:
		return t.DoubleValue, nil
	case *policy.Value_DurationValue:
		// milliseconds is the NR preferred unit
		return float64(t.DurationValue.Value.Seconds)*1000.0 + float64(t.DurationValue.Value.Nanos)*0.000001, nil
	default:
		return 0.0, fmt.Errorf("unknown float64 conversion for %v (%T)", t, t)
	}
	panic("fell through metric value float64 parsing")
}

// DimensionsToAttributes returns an appropriate set of New Relic attributes from the passed Istio telemetry dimensions.
func DimensionsToAttributes(in map[string]*policy.Value) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = ValueToAttribute(v)
	}
	return out
}

// ValueToAttribute returns metric attribute values as types the telemetry-sdk handles:
//	   timestamps and durations       -> milliseconds as float64
//	   string, bool, all number types -> unchanged
//	   everything else                -> string
func ValueToAttribute(in *policy.Value) interface{} {
	switch t := in.GetValue().(type) {
	case *policy.Value_StringValue:
		return t.StringValue
	case *policy.Value_Int64Value:
		return t.Int64Value
	case *policy.Value_DoubleValue:
		return t.DoubleValue
	case *policy.Value_BoolValue:
		return t.BoolValue
	case *policy.Value_DurationValue:
		return float64(t.DurationValue.Value.Seconds)*1000.0 + float64(t.DurationValue.Value.Nanos)*0.000001 // milliseconds NR preferred unit
	case *policy.Value_TimestampValue:
		return float64(t.TimestampValue.Value.Seconds) * 1000.0
	case *policy.Value_IpAddressValue:
		return adapter.Stringify(t.IpAddressValue.Value)
	case *policy.Value_EmailAddressValue:
		return adapter.Stringify(t.EmailAddressValue.Value)
	case *policy.Value_DnsNameValue:
		return adapter.Stringify(t.DnsNameValue.Value)
	case *policy.Value_UriValue:
		return adapter.Stringify(t.UriValue.Value)
	case *policy.Value_StringMapValue:
		return adapter.Stringify(t.StringMapValue.Value)
	default:
		return fmt.Sprintf("%v", in)
	}
	panic("fell through attribute value parsing")
}
