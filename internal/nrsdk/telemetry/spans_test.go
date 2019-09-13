package telemetry

import (
	"bytes"
	"testing"
	"time"
)

func BenchmarkSpansJSON(b *testing.B) {
	// This benchmark tests the overhead of turning spans into JSON.
	batch := &spanBatch{}
	numSpans := 10 * 1000
	tm := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)

	for i := 0; i < numSpans; i++ {
		batch.AddSpan(Span{
			ID:          "myid",
			TraceID:     "mytraceid",
			Name:        "myname",
			ParentID:    "myparent",
			Timestamp:   tm,
			Duration:    2 * time.Second,
			ServiceName: "myentity",
			Attributes: map[string]interface{}{
				"zip": "zap",
				"zop": 123,
			},
		})
	}

	if len(batch.Spans) != numSpans {
		b.Fatal(len(batch.Spans))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		batch.writeJSON(buf)
		if bts := buf.Bytes(); nil == bts || len(bts) == 0 {
			b.Fatal(string(bts))
		}
	}
}

func testHarvesterSpans(t testing.TB, h *Harvester, expect string) {
	reqs := h.swapOutSpans()
	if nil == reqs {
		if expect != "null" {
			t.Error("nil spans", expect)
		}
		return
	}
	if len(reqs) != 1 {
		t.Fatal(reqs)
	}
	js := reqs[0].UncompressedBody
	actual := string(js)
	if th, ok := t.(interface{ Helper() }); ok {
		th.Helper()
	}
	compactExpect := compactJSONString(expect)
	if compactExpect != actual {
		t.Errorf("\nexpect=%s\nactual=%s\n", compactExpect, actual)
	}
}

func TestSpan(t *testing.T) {
	tm := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	h := NewHarvester(configTesting)
	h.RecordSpan(Span{
		ID:          "myid",
		TraceID:     "mytraceid",
		Name:        "myname",
		ParentID:    "myparent",
		Timestamp:   tm,
		Duration:    2 * time.Second,
		ServiceName: "myentity",
		Attributes: map[string]interface{}{
			"zip": "zap",
		},
	})
	expect := `[{"common":{},"spans":[{
		"id":"myid",
		"trace.id":"mytraceid",
		"timestamp":1417136460000,
		"attributes": {
			"name":"myname",
			"parent.id":"myparent",
			"duration.ms":2000,
			"service.name":"myentity",
			"zip":"zap"
		}
	}]}]`
	testHarvesterSpans(t, h, expect)
}

func TestSpanInvalidAttribute(t *testing.T) {
	tm := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	h := NewHarvester(configTesting)
	h.RecordSpan(Span{
		ID:          "myid",
		TraceID:     "mytraceid",
		Name:        "myname",
		ParentID:    "myparent",
		Timestamp:   tm,
		Duration:    2 * time.Second,
		ServiceName: "myentity",
		Attributes: map[string]interface{}{
			"weird-things-get-turned-to-strings": struct{}{},
			"nil-gets-removed":                   nil,
		},
	})
	expect := `[{"common":{},"spans":[{
		"id":"myid",
		"trace.id":"mytraceid",
		"timestamp":1417136460000,
		"attributes": {
			"name":"myname",
			"parent.id":"myparent",
			"duration.ms":2000,
			"service.name":"myentity",
			"weird-things-get-turned-to-strings":"struct {}"
		}
	}]}]`
	testHarvesterSpans(t, h, expect)
}

func TestNoAPIKeyNoSpan(t *testing.T) {
	tm := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	h := NewHarvester()
	err := h.RecordSpan(Span{
		ID:          "myid",
		TraceID:     "mytraceid",
		Name:        "myname",
		ParentID:    "myparent",
		Timestamp:   tm,
		Duration:    2 * time.Second,
		ServiceName: "myentity",
		Attributes: map[string]interface{}{
			"zip": "zap",
			"zop": 123,
		},
	})
	if err != errSpansDisabled {
		t.Error(err)
	}
	if 0 != len(h.spans) {
		t.Error("spans were recorded", h.spans)
	}

	expect := "null"
	testHarvesterSpans(t, h, expect)
}
