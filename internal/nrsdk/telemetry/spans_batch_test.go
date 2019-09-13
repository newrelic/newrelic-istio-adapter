package telemetry

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/internal"
)

func testSpanBatchJSON(t testing.TB, batch *spanBatch, expect string) {
	if th, ok := t.(interface{ Helper() }); ok {
		th.Helper()
	}
	reqs, err := batch.NewRequests("licensKey", "")
	if nil != err {
		t.Fatal(err)
	}
	if len(reqs) != 1 {
		t.Fatal(reqs)
	}
	req := reqs[0]
	actual := string(req.UncompressedBody)
	compact := compactJSONString(expect)
	if actual != compact {
		t.Errorf("\nexpect=%s\nactual=%s\n", compact, actual)
	}

	body, err := ioutil.ReadAll(req.Request.Body)
	req.Request.Body.Close()
	if err != nil {
		t.Fatal("unable to read body", err)
	}
	if len(body) != req.compressedBodyLength {
		t.Error("compressed body length mismatch",
			len(body), req.compressedBodyLength)
	}
	uncompressed, err := internal.Uncompress(body)
	if err != nil {
		t.Fatal("unable to uncompress body", err)
	}
	if string(uncompressed) != string(req.UncompressedBody) {
		t.Error("request JSON mismatch", string(uncompressed), string(req.UncompressedBody))
	}
}

func TestSpansPayloadSplit(t *testing.T) {
	// test len 0
	sp := &spanBatch{}
	split := sp.split()
	if split != nil {
		t.Error(split)
	}

	// test len 1
	sp = &spanBatch{Spans: []Span{{Name: "a"}}}
	split = sp.split()
	if split != nil {
		t.Error(split)
	}

	// test len 2
	sp = &spanBatch{Spans: []Span{{Name: "a"}, {Name: "b"}}}
	split = sp.split()
	if len(split) != 2 {
		t.Error("split into incorrect number of slices", len(split))
	}
	testSpanBatchJSON(t, split[0].(*spanBatch), `[{"common":{},"spans":[{"id":"","trace.id":"","timestamp":-6795364578871,"attributes":{"name":"a"}}]}]`)
	testSpanBatchJSON(t, split[1].(*spanBatch), `[{"common":{},"spans":[{"id":"","trace.id":"","timestamp":-6795364578871,"attributes":{"name":"b"}}]}]`)

	// test len 3
	sp = &spanBatch{Spans: []Span{{Name: "a"}, {Name: "b"}, {Name: "c"}}}
	split = sp.split()
	if len(split) != 2 {
		t.Error("split into incorrect number of slices", len(split))
	}
	testSpanBatchJSON(t, split[0].(*spanBatch), `[{"common":{},"spans":[{"id":"","trace.id":"","timestamp":-6795364578871,"attributes":{"name":"a"}}]}]`)
	testSpanBatchJSON(t, split[1].(*spanBatch), `[{"common":{},"spans":[{"id":"","trace.id":"","timestamp":-6795364578871,"attributes":{"name":"b"}},{"id":"","trace.id":"","timestamp":-6795364578871,"attributes":{"name":"c"}}]}]`)
}

func TestSpansJSON(t *testing.T) {
	batch := &spanBatch{Spans: []Span{
		{}, // Empty span
		{ // Span with everything
			ID:          "myid",
			TraceID:     "mytraceid",
			Name:        "myname",
			ParentID:    "myparentid",
			Timestamp:   time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
			Duration:    2 * time.Second,
			ServiceName: "myentity",
			Attributes:  map[string]interface{}{"zip": "zap"},
		},
	}}
	testSpanBatchJSON(t, batch, `[{"common":{},"spans":[
		{
			"id":"",
			"trace.id":"",
			"timestamp":-6795364578871,
			"attributes": {
			}
		},
		{
			"id":"myid",
			"trace.id":"mytraceid",
			"timestamp":1417136460000,
			"attributes": {
				"name":"myname",
				"parent.id":"myparentid",
				"duration.ms":2000,
				"service.name":"myentity",
				"zip":"zap"
			}
		}
	]}]`)
}

func TestSpansJSONWithCommonAttributes(t *testing.T) {
	batch := &spanBatch{
		Attributes: map[string]interface{}{"zup": "wup"},
		Spans: []Span{
			{
				ID:          "myid",
				TraceID:     "mytraceid",
				Name:        "myname",
				ParentID:    "myparentid",
				Timestamp:   time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
				Duration:    2 * time.Second,
				ServiceName: "myentity",
				Attributes:  map[string]interface{}{"zip": "zap"},
			},
		}}
	testSpanBatchJSON(t, batch, `[{"common":{"attributes":{"zup":"wup"}},"spans":[
		{
			"id":"myid",
			"trace.id":"mytraceid",
			"timestamp":1417136460000,
			"attributes": {
				"name":"myname",
				"parent.id":"myparentid",
				"duration.ms":2000,
				"service.name":"myentity",
				"zip":"zap"
			}
		}
	]}]`)
}

func TestSpansJSONWithCommonAttributesJSON(t *testing.T) {
	batch := &spanBatch{
		AttributesJSON: json.RawMessage(`{"zup":"wup"}`),
		Spans: []Span{
			{
				ID:          "myid",
				TraceID:     "mytraceid",
				Name:        "myname",
				ParentID:    "myparentid",
				Timestamp:   time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC),
				Duration:    2 * time.Second,
				ServiceName: "myentity",
				Attributes:  map[string]interface{}{"zip": "zap"},
			},
		}}
	testSpanBatchJSON(t, batch, `[{"common":{"attributes":{"zup":"wup"}},"spans":[
		{
			"id":"myid",
			"trace.id":"mytraceid",
			"timestamp":1417136460000,
			"attributes": {
				"name":"myname",
				"parent.id":"myparentid",
				"duration.ms":2000,
				"service.name":"myentity",
				"zip":"zap"
			}
		}
	]}]`)
}

func TestSpanRequestURL(t *testing.T) {
	batch := &spanBatch{Spans: []Span{{}}}

	// Default URL
	reqs, err := batch.NewRequests("key", "")
	if err != nil {
		t.Error(err)
	}
	if len(reqs) != 1 {
		t.Fatal(reqs)
	}
	req := reqs[0]
	if req.Request.URL.String() != "https://trace-api.newrelic.com/trace/v1" {
		t.Error(req.Request.URL.String())
	}
	// Override URL
	reqs, err = batch.NewRequests("key", "https://override.host.com/path")
	if err != nil {
		t.Error(err)
	}
	if len(reqs) != 1 {
		t.Fatal(reqs)
	}
	req = reqs[0]
	if req.Request.URL.String() != "https://override.host.com/path" {
		t.Error(req.Request.URL.String())
	}
}
