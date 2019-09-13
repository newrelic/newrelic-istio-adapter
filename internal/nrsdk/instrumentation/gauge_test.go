package instrumentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"testing"
	"time"
)

// compactJSONString removes the whitespace from a JSON string.  This function
// will panic if the string provided is not valid JSON.
func compactJSONString(js string) string {
	buf := new(bytes.Buffer)
	if err := json.Compact(buf, []byte(js)); err != nil {
		panic(fmt.Errorf("unable to compact JSON: %v", err))
	}
	return buf.String()
}

// sortedMetricsHelper is used to sort metrics for JSON comparison.
type sortedMetricsHelper []json.RawMessage

func (h sortedMetricsHelper) Len() int {
	return len(h)
}
func (h sortedMetricsHelper) Less(i, j int) bool {
	return string(h[i]) < string(h[j])
}
func (h sortedMetricsHelper) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func testMetrics(t testing.TB, ag *MetricAggregator, expect string) {
	js, err := json.Marshal(ag.Metrics())
	if err != nil {
		t.Fatal(err)
	}
	var helper sortedMetricsHelper
	if err := json.Unmarshal(js, &helper); err != nil {
		t.Fatal("unable to unmarshal metrics for sorting", err)
		return
	}
	sort.Sort(helper)
	js, err = json.Marshal(helper)
	if nil != err {
		t.Fatal("unable to marshal metrics", err)
		return
	}
	actual := string(js)

	if th, ok := t.(interface{ Helper() }); ok {
		th.Helper()
	}
	compactExpect := compactJSONString(expect)
	if compactExpect != actual {
		t.Errorf("\nexpect=%s\nactual=%s\n", compactExpect, actual)
	}
}

func TestGauge(t *testing.T) {
	now := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	ag := NewMetricAggregator()
	ag.NewGauge("myGauge", map[string]interface{}{"zip": "zap"}).valueNow(123.4, now)

	expect := `[{"Name":"myGauge","Attributes":null,"AttributesJSON":{"zip":"zap"},"Value":123.4,"Timestamp":"2014-11-28T01:01:00Z"}]`
	testMetrics(t, ag, expect)
}

func TestNilAggregatorGauges(t *testing.T) {
	now := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	var ag *MetricAggregator
	gauge := ag.NewGauge("gauge", map[string]interface{}{})
	gauge.valueNow(5.5, now)
}

func TestNilGaugeMethods(t *testing.T) {
	now := time.Date(2014, time.November, 28, 1, 1, 0, 0, time.UTC)
	var gauge *Gauge
	gauge.valueNow(5.5, now)
}

func TestGaugeNilAggregator(t *testing.T) {
	g := Gauge{}
	g.Value(10)
}
