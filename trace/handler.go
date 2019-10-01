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

package trace

import (
	"context"
	"errors"

	"github.com/gogo/protobuf/types"
	"github.com/newrelic/newrelic-istio-adapter/convert"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"istio.io/istio/mixer/pkg/adapter"
	"istio.io/istio/mixer/template/tracespan"
)

// Handler represents a processor that can handle tracespans from Istio and transmit them to New Relic.
type Handler struct {
	logger    adapter.Logger
	harvester *telemetry.Harvester
}

// HandleTraceSpan transforms tracespan template instances into New Relic spans and
// sends them to New Relic.
func (h *Handler) HandleTraceSpan(_ context.Context, msgs []*tracespan.InstanceMsg) error {
	for _, i := range msgs {
		span, err := convertTraceSpan(i)
		if err != nil {
			h.logger.Warningf("Error converting tracespan: %v", err)
			continue
		}

		h.harvester.RecordSpan(*span)
	}
	return nil
}

// convertTraceSpan will convert a tracespan.InstanceMsg into a telemetry.Span.
func convertTraceSpan(i *tracespan.InstanceMsg) (*telemetry.Span, error) {
	startTime, err := types.TimestampFromProto(i.StartTime.GetValue())
	if err != nil {
		return nil, err
	}
	endTime, err := types.TimestampFromProto(i.EndTime.GetValue())
	if err != nil {
		return nil, err
	}

	if i.TraceId == "" {
		return nil, errors.New("no trace ID")
	}

	if i.SpanId == "" {
		return nil, errors.New("no span ID")
	}

	attributes := convert.DimensionsToAttributes(i.SpanTags)
	attributes["response.code"] = i.HttpStatusCode
	attributes["clientSpan"] = i.ClientSpan
	attributes["rewriteClientSpanId"] = i.RewriteClientSpanId
	attributes["source.name"] = i.SourceName
	attributes["source.ip"] = adapter.Stringify(i.SourceIp.GetValue())
	attributes["destination.name"] = i.DestinationName
	attributes["destination.ip"] = adapter.Stringify(i.DestinationIp.GetValue())
	attributes["request.size"] = i.RequestSize
	attributes["request.totalSize"] = i.RequestTotalSize
	attributes["response.size"] = i.ResponseSize
	attributes["response.totalSize"] = i.ResponseTotalSize
	attributes["api.protocol"] = i.ApiProtocol

	span := &telemetry.Span{
		ID:        i.SpanId,
		TraceID:   i.TraceId,
		Name:      i.SpanName,
		ParentID:  i.ParentSpanId,
		Timestamp: startTime,
		Duration:  endTime.Sub(startTime),
		// Default to assuming this was a server span.
		ServiceName: i.DestinationName,
		Attributes:  attributes,
	}

	if i.ClientSpan {
		// For explicit client spans, assign to the source.
		span.ServiceName = i.SourceName
	}

	return span, nil
}
