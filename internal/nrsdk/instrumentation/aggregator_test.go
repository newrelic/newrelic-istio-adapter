package instrumentation

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/internal"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
)

func TestDifferentAttributes(t *testing.T) {
	// Test that attributes contribute to identity, ie, metrics with the
	// same name but different attributes should generate different metrics.
	now := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	ag := NewMetricAggregator()
	ag.NewGauge("myGauge", map[string]interface{}{"zip": "zap"}).valueNow(1.0, now)
	ag.NewGauge("myGauge", map[string]interface{}{"zip": "zup"}).valueNow(2.0, now)
	expect := `[
		{"Name":"myGauge","Attributes":null,"AttributesJSON":{"zip":"zap"},"Value":1,"Timestamp":"2014-11-28T01:01:00Z"},
		{"Name":"myGauge","Attributes":null,"AttributesJSON":{"zip":"zup"},"Value":2,"Timestamp":"2014-11-28T01:01:00Z"}
	]`
	testMetrics(t, ag, expect)
}

func TestSameNameDifferentTypes(t *testing.T) {
	// Test that type contributes to identity, ie, metrics with the same
	// name and same attributes of different types should generate different
	// metrics.
	ag := NewMetricAggregator()
	now := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	ag.NewGauge("metric", map[string]interface{}{"zip": "zap"}).valueNow(1.0, now)
	ag.NewCount("metric", map[string]interface{}{"zip": "zap"}).Increment()
	ag.NewSummary("metric", map[string]interface{}{"zip": "zap"}).Record(1.0)
	expect := `[
		{"Name":"metric","Attributes":null,"AttributesJSON":{"zip":"zap"},"Count":1,"Sum":1,"Min":1,"Max":1,"Timestamp":"0001-01-01T00:00:00Z","Interval":0},
		{"Name":"metric","Attributes":null,"AttributesJSON":{"zip":"zap"},"Value":1,"Timestamp":"0001-01-01T00:00:00Z","Interval":0},
		{"Name":"metric","Attributes":null,"AttributesJSON":{"zip":"zap"},"Value":1,"Timestamp":"2014-11-28T01:01:00Z"}
	]`
	testMetrics(t, ag, expect)
}

func TestManyAttributes(t *testing.T) {
	// Test adding the same metric with many attributes to ensure that
	// attributes are serialized into JSON in a fixed order.  Note that if
	// JSON attribute order is random this test may still occasionally pass.
	now := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	ag := NewMetricAggregator()
	attributes := make(map[string]interface{})
	for i := 0; i < 100; i++ {
		attributes[strconv.Itoa(i)] = i
	}
	ag.NewGauge("myGauge", attributes).valueNow(1.0, now)
	ag.NewGauge("myGauge", attributes).valueNow(2.0, now)
	if len(ag.metrics) != 1 {
		t.Fatal(len(ag.metrics))
	}
}

func TestBeforeHarvestFunc(t *testing.T) {
	ag := NewMetricAggregator()

	cfg := &telemetry.Config{}
	ag.BeforeHarvest(cfg)
	if nil == cfg.BeforeHarvestFunc {
		t.Error("BeforeHarvestFunc not set")
	}

	var calls int
	cfg = &telemetry.Config{
		BeforeHarvestFunc: func(h *telemetry.Harvester) {
			calls++
		},
	}
	ag.BeforeHarvest(cfg)
	if nil == cfg.BeforeHarvestFunc {
		t.Error("BeforeHarvestFunc not set")
	}
	cfg.BeforeHarvestFunc(nil)
	if 1 != calls {
		t.Error("original BeforeHarvestFunc not called")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

// optional interface required for go1.4 and go1.5
func (fn roundTripperFunc) CancelRequest(*http.Request) {}

func emptyResponse(status int) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(""))),
	}
}

func TestBeforeHarvestFuncRecordsMetrics(t *testing.T) {
	ag := NewMetricAggregator()
	ag.NewCount("mycount", nil).Increment()

	var posts int
	rt := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		posts++
		body, err := ioutil.ReadAll(req.Body)
		if nil != err {
			t.Fatal(err)
		}
		defer req.Body.Close()
		uncompressed, err := internal.Uncompress(body)
		if err != nil {
			t.Fatal(err)
		}
		var helper []struct {
			Metrics json.RawMessage `json:"metrics"`
		}
		if err := json.Unmarshal(uncompressed, &helper); err != nil {
			t.Fatal("unable to unmarshal metrics for sorting", err)
		}

		expected := `[{"name":"mycount","type":"count","value":1,"attributes":{}}]`
		if string(helper[0].Metrics) != expected {
			t.Error("incorrect metrics found", helper[0].Metrics)
		}

		return emptyResponse(200), nil
	})

	h := telemetry.NewHarvester(
		ag.BeforeHarvest,
		func(cfg *telemetry.Config) {
			cfg.APIKey = "api-key"
			cfg.HarvestPeriod = 0
			cfg.Client.Transport = rt
		},
	)
	h.HarvestNow(context.Background())
	if 1 != posts {
		t.Error("no metric data posted")
	}
}
