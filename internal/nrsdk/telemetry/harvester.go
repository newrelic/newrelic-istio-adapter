package telemetry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/internal"
)

// Harvester aggregates and reports metrics and spans.
type Harvester struct {
	// These fields are not modified after Harvester creation.  They may be
	// safely accessed without locking.
	config               Config
	commonAttributesJSON json.RawMessage

	// lock protects the mutable fields below.
	lock        sync.Mutex
	lastHarvest time.Time
	rawMetrics  []Metric
	spans       []Span
}

const (
	// NOTE:  These constant values are used in Config field doc comments.
	defaultHarvestPeriod  = 5 * time.Second
	defaultRetryBackoff   = 3 * time.Second
	defaultHarvestTimeout = 15 * time.Second
)

// NewHarvester creates a new harvester.
func NewHarvester(options ...func(*Config)) *Harvester {
	cfg := Config{
		Client:         &http.Client{},
		HarvestPeriod:  defaultHarvestPeriod,
		HarvestTimeout: defaultHarvestTimeout,
		RetryBackoff:   defaultRetryBackoff,
	}
	for _, opt := range options {
		opt(&cfg)
	}

	h := &Harvester{
		config:      cfg,
		lastHarvest: time.Now(),
	}

	// Marshal the common attributes to JSON here to avoid doing it on every
	// harvest.  This also has the benefit that it avoids race conditions if
	// the consumer modifies the CommonAttributes map after calling
	// NewHarvester.
	if nil != h.config.CommonAttributes {
		attrs := vetAttributes(h.config.CommonAttributes, h.config.logError)
		attributesJSON, err := json.Marshal(attrs)
		if err != nil {
			h.config.logError(map[string]interface{}{
				"err":     err.Error(),
				"message": "error marshaling common attributes",
			})
		} else {
			h.commonAttributesJSON = attributesJSON
		}
		h.config.CommonAttributes = nil
	}

	spawnHarvester := h.needsHarvestThread()

	h.config.logDebug(map[string]interface{}{
		"event":                   "harvester created",
		"api-key":                 h.config.APIKey,
		"harvest-period-seconds":  h.config.HarvestPeriod.Seconds(),
		"spawn-harvest-goroutine": spawnHarvester,
		"metrics-url-override":    h.config.MetricsURLOverride,
		"spans-url-override":      h.config.SpansURLOverride,
		"collect-metrics":         h.collectMetrics(),
		"collect-spans":           h.collectSpans(),
		"version":                 internal.Version,
	})

	if spawnHarvester {
		go h.harvest()
	}

	return h
}

func (h *Harvester) needsHarvestThread() bool {
	if 0 == h.config.HarvestPeriod {
		return false
	}
	if !h.collectMetrics() && !h.collectSpans() {
		return false
	}
	return true
}

func (h *Harvester) collectMetrics() bool {
	if nil == h {
		return false
	}
	if "" == h.config.APIKey {
		return false
	}
	return true
}

func (h *Harvester) collectSpans() bool {
	if nil == h {
		return false
	}
	if "" == h.config.APIKey {
		return false
	}
	return true
}

var (
	errSpanIDUnset   = errors.New("span id must be set")
	errTraceIDUnset  = errors.New("trace id must be set")
	errSpansDisabled = errors.New("spans are not enabled: APIKey unset")
)

// RecordSpan records the given span.
func (h *Harvester) RecordSpan(s Span) error {
	if nil == h {
		return nil
	}
	if "" == h.config.APIKey {
		return errSpansDisabled
	}
	if "" == s.TraceID {
		return errTraceIDUnset
	}
	if "" == s.ID {
		return errSpanIDUnset
	}
	if s.Timestamp.IsZero() {
		s.Timestamp = time.Now()
	}

	// TODO: Remove Attributes which collide with the recommended
	// attributes.

	h.lock.Lock()
	defer h.lock.Unlock()

	h.spans = append(h.spans, s)
	return nil
}

// RecordMetric adds a fully formed metric.  This metric is not aggregated with
// any other metrics and is never dropped.  The Timestamp field must be
// specified on Gauge metrics.  The Timestamp/Interval fields on Count and
// Summary are optional and will be assumed to be the harvester batch times if
// unset.
func (h *Harvester) RecordMetric(m Metric) {
	if !h.collectMetrics() {
		return
	}
	h.lock.Lock()
	defer h.lock.Unlock()

	h.rawMetrics = append(h.rawMetrics, m)
}

type response struct {
	statusCode int
	body       []byte
	err        error
	retryAfter string
}

func (r response) needsRetry(cfg *Config) (bool, time.Duration) {
	switch r.statusCode {
	case 202, 200:
		// success
		return false, 0
	case 400, 403, 404, 405, 411, 413:
		// errors that should not retry
		return false, 0
	case 429:
		// special retry backoff time
		if "" != r.retryAfter {
			// Honor Retry-After header value in seconds
			if d, err := time.ParseDuration(r.retryAfter + "s"); nil == err {
				if d > cfg.RetryBackoff {
					return true, d
				}
			}
		}
		return true, cfg.RetryBackoff
	default:
		// all other errors should retry
		return true, cfg.RetryBackoff
	}
}

func postData(req *http.Request, client *http.Client) response {
	resp, err := client.Do(req)
	if nil != err {
		return response{err: fmt.Errorf("error posting data: %v", err)}
	}
	defer resp.Body.Close()

	r := response{
		statusCode: resp.StatusCode,
		retryAfter: resp.Header.Get("Retry-After"),
	}

	// On success, metrics ingest returns 202, span ingest returns 200.
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		r.body, _ = ioutil.ReadAll(resp.Body)
	} else {
		r.err = fmt.Errorf("unexpected post response code: %d: %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	return r
}

func (h *Harvester) swapOutMetrics(now time.Time) []request {
	if !h.collectMetrics() {
		return nil
	}

	h.lock.Lock()
	lastHarvest := h.lastHarvest
	h.lastHarvest = now
	rawMetrics := h.rawMetrics
	h.rawMetrics = nil
	h.lock.Unlock()

	if 0 == len(rawMetrics) {
		return nil
	}

	batch := &metricBatch{
		Timestamp:      lastHarvest,
		Interval:       now.Sub(lastHarvest),
		AttributesJSON: h.commonAttributesJSON,
		Metrics:        rawMetrics,
	}
	reqs, err := batch.NewRequests(h.config.APIKey, h.config.MetricsURLOverride)
	if nil != err {
		h.config.logError(map[string]interface{}{
			"err":     err.Error(),
			"message": "error creating requests for metrics",
		})
		return nil
	}
	return reqs
}

func (h *Harvester) swapOutSpans() []request {
	if !h.collectSpans() {
		return nil
	}

	h.lock.Lock()
	sps := h.spans
	h.spans = nil
	h.lock.Unlock()

	if nil == sps {
		return nil
	}
	batch := &spanBatch{
		AttributesJSON: h.commonAttributesJSON,
		Spans:          sps,
	}
	reqs, err := batch.NewRequests(h.config.APIKey, h.config.SpansURLOverride)
	if nil != err {
		h.config.logError(map[string]interface{}{
			"err":     err.Error(),
			"message": "error creating requests for spans",
		})
		return nil
	}
	return reqs
}

func harvestRequest(req request, cfg *Config) {
	for {
		cfg.logDebug(map[string]interface{}{
			"event":       "data post",
			"url":         req.Request.URL.String(),
			"body-length": req.compressedBodyLength,
		})
		// Check if the audit log is enabled to prevent unnecessarily
		// copying UncompressedBody.
		if cfg.auditLogEnabled() {
			cfg.logAudit(map[string]interface{}{
				"event": "uncompressed request body",
				"url":   req.Request.URL.String(),
				"data":  jsonString(req.UncompressedBody),
			})
		}

		resp := postData(req.Request, cfg.Client)

		if nil != resp.err {
			cfg.logError(map[string]interface{}{
				"err": resp.err.Error(),
			})
		} else {
			cfg.logDebug(map[string]interface{}{
				"event":  "data post response",
				"status": resp.statusCode,
				"body":   jsonOrString(resp.body),
			})
		}
		retry, backoff := resp.needsRetry(cfg)
		if !retry {
			return
		}

		tmr := time.NewTimer(backoff)
		select {
		case <-tmr.C:
			continue
		case <-req.Request.Context().Done():
			tmr.Stop()
			return
		}
	}
}

// HarvestNow sends metric and span data to New Relic.  This method blocks until
// all data has been sent successfully or the Config.HarvestTimeout timeout has
// elapsed. This method can be used with a zero Config.HarvestPeriod value to
// control exactly when data is sent to New Relic servers.
func (h *Harvester) HarvestNow(ct context.Context) {
	if nil == h {
		return
	}

	ctx, cancel := context.WithTimeout(ct, h.config.HarvestTimeout)
	defer cancel()

	if nil != h.config.BeforeHarvestFunc {
		h.config.BeforeHarvestFunc(h)
	}

	var reqs []request
	reqs = append(reqs, h.swapOutMetrics(time.Now())...)
	reqs = append(reqs, h.swapOutSpans()...)

	for _, req := range reqs {
		req.Request = req.Request.WithContext(ctx)
		harvestRequest(req, &h.config)
		if err := ctx.Err(); err != nil {
			// NOTE: It is possible that the context was
			// cancelled/timedout right after the request
			// successfully finished.  In that case, we will
			// erroneously log a message.  I (will) don't think
			// that's worth trying to engineer around.
			h.config.logError(map[string]interface{}{
				"event":         "harvest cancelled or timed out",
				"message":       "dropping data",
				"context-error": err.Error(),
			})
			return
		}
	}
}

// harvest concurrently harvests telemetry data
func (h *Harvester) harvest() {
	ticker := time.NewTicker(h.config.HarvestPeriod)
	for range ticker.C {
		go h.HarvestNow(context.Background())
	}
}
