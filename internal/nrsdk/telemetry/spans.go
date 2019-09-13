package telemetry

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/internal"
)

// Span is a distributed tracing span.
type Span struct {
	// Required Fields:
	//
	// ID is a unique identifier for this span.
	ID string
	// TraceID is a unique identifier shared by all spans within a single
	// trace.
	TraceID string
	// Timestamp is when this span started.  If Timestamp is not set, it
	// will be assigned to time.Now() in Harvester.RecordSpan.
	Timestamp time.Time

	// Recommended Fields:
	//
	// Name is the name of this span.
	Name string
	// ParentID is the span id of the previous caller of this span.  This
	// can be empty if this is the first span.
	ParentID string
	// Duration is the duration of this span.  This field will be reported
	// in milliseconds.
	Duration time.Duration
	// ServiceName is the name of the service that created this span.
	ServiceName string

	// Additional Fields:
	//
	// Attributes is a map of user specified tags on this span.  The map
	// values can be any of bool, number, or string.
	Attributes map[string]interface{}
	// AttributesJSON is a json.RawMessage of attributes for this metric. It
	// will only be sent if Attributes is nil.
	AttributesJSON json.RawMessage
}

func (s *Span) writeJSON(buf *bytes.Buffer) {
	w := internal.JSONFieldsWriter{Buf: buf}
	buf.WriteByte('{')

	w.StringField("id", s.ID)
	w.StringField("trace.id", s.TraceID)
	w.IntField("timestamp", s.Timestamp.UnixNano()/(1000*1000))

	w.AddKey("attributes")
	buf.WriteByte('{')
	ww := internal.JSONFieldsWriter{Buf: buf}

	if "" != s.Name {
		ww.StringField("name", s.Name)
	}
	if "" != s.ParentID {
		ww.StringField("parent.id", s.ParentID)
	}
	if 0 != s.Duration {
		ww.FloatField("duration.ms", s.Duration.Seconds()*1000.0)
	}
	if "" != s.ServiceName {
		ww.StringField("service.name", s.ServiceName)
	}

	internal.AddAttributes(&ww, s.Attributes)

	buf.WriteByte('}')
	buf.WriteByte('}')
}

// spanBatch represents a single batch of spans to report to New Relic.
type spanBatch struct {
	// Attributes is a map of attributes to apply to all spans in this
	// spanBatch. They are included in addition to any attributes set on
	// any particular span.
	Attributes map[string]interface{}
	// AttributesJSON is a json.RawMessage of attributes to apply to all
	// spans in this spanBatch. It will only be sent if the Attributes field
	// on this spanBatch is nil. These attributes are included in addition
	// to any attributes on any particular span.
	AttributesJSON json.RawMessage
	// TODO: Add ID and TraceID here.
	Spans []Span
}

// AddSpan appends a span to the spanBatch.
func (batch *spanBatch) AddSpan(s Span) {
	batch.Spans = append(batch.Spans, s)
}

// split will split the spanBatch into 2 equally sized batches.
// If the number of spans in the original is 0 or 1 then nil is returned.
func (batch *spanBatch) split() []requestsBuilder {
	if len(batch.Spans) < 2 {
		return nil
	}

	half := len(batch.Spans) / 2
	b1 := *batch
	b1.Spans = batch.Spans[:half]
	b2 := *batch
	b2.Spans = batch.Spans[half:]

	return []requestsBuilder{
		requestsBuilder(&b1),
		requestsBuilder(&b2),
	}
}

func (batch *spanBatch) writeJSON(buf *bytes.Buffer) {
	buf.WriteByte('[')
	buf.WriteByte('{')
	w := internal.JSONFieldsWriter{Buf: buf}

	w.AddKey("common")
	buf.WriteByte('{')
	ww := internal.JSONFieldsWriter{Buf: buf}
	if nil != batch.Attributes {
		ww.WriterField("attributes", internal.Attributes(batch.Attributes))
	} else if nil != batch.AttributesJSON {
		ww.RawField("attributes", batch.AttributesJSON)
	}
	buf.WriteByte('}')

	w.AddKey("spans")
	buf.WriteByte('[')
	for idx, s := range batch.Spans {
		if idx > 0 {
			buf.WriteByte(',')
		}
		s.writeJSON(buf)
	}
	buf.WriteByte(']')
	buf.WriteByte('}')
	buf.WriteByte(']')
}

const (
	defaultSpanURL = "https://trace-api.newrelic.com/trace/v1"
)

// NewRequests creates new requests from the spanBatch. The request can be
// sent with an http.Client.
//
// NewRequest returns requests or an error if there was one.  Each Request
// has an UncompressedBody field that is useful in debugging or testing.
//
// Possible response codes to be expected when sending the request to New
// Relic:
//
//  202 for success
//  403 for an auth failure
//  404 for a bad path
//  405 for anything but POST
//  411 if the Content-Length header is not included
//  413 for a payload that is too large
//  400 for a generally invalid request
//  429 Too Many Requests
//
func (batch *spanBatch) NewRequests(apiKey, urlOverride string) ([]request, error) {
	return newRequests(batch, apiKey, urlOverride, maxCompressedSizeBytes)
}

func (batch *spanBatch) newRequest(apiKey, urlOverride string) (request, error) {
	buf := &bytes.Buffer{}
	batch.writeJSON(buf)
	return newRequest(defaultSpanURL, urlOverride, apiKey, buf.Bytes())
}
