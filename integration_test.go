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

// +build integration

package newrelic

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/instrumentation"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
	integration "istio.io/istio/mixer/pkg/adapter/test"
)

const (
	adapterConfigPath  = "config/newrelic.yaml"
	operatorConfigPath = "integration_test_cfg.yaml"
)

func TestReport(t *testing.T) {
	adapterCfg, err := ioutil.ReadFile(adapterConfigPath)
	if err != nil {
		t.Fatalf("failed to read adapter config file: %v", err)
	}

	operatorCfg, err := ioutil.ReadFile(operatorConfigPath)
	if err != nil {
		t.Fatalf("failed to read operator config file: %v", err)
	}

	var commonAttrs = map[string]interface{}{
		"cluster.name":      "hotdog-stand",
		"foodHandlerPermit": "revoked",
	}
	mockt := newMockTransport()
	agg, harvester := mockHarvester(mockt, commonAttrs)

	scenario := integration.Scenario{
		Setup: func() (ctx interface{}, err error) {
			s, err := NewServer(":0", agg, harvester)
			if err != nil {
				return nil, err
			}

			// start the adapter
			go func() {
				s.Run()
			}()

			return s, nil
		},

		Teardown: func(ctx interface{}) {
			s := ctx.(*Server)
			s.Close()
		},

		ParallelCalls: []integration.Call{
			{
				CallKind: integration.REPORT,
				Attrs: map[string]interface{}{
					"context.protocol":          "http",
					"destination.workload.name": "cat",
					"source.labels":             map[string]string{"app": "grocery"},
					"response.code":             int64(418),
					"request.time":              time.Unix(1582230020, 0),
					"source.ip":                 []byte(net.ParseIP("1.2.3.4")),
					"connection.duration":       127 * time.Hour,
					"connection.mtls":           true,
				},
			},
		},

		GetState: func(ctx interface{}) (interface{}, error) {
			// send metrics immediately
			harvester.HarvestNow(context.Background())

			// verify a single request was issued with the correct metric data
			if len(mockt.requests) != 1 {
				t.Fatalf("expected a single request to New Relic metric api, %d requests reported",
					len(mockt.requests))
			}

			common, metrics, err := readTelemetryRequestJson(mockt.requests[0])
			if err != nil {
				t.Fatalf("failed to unmarshal the telemetry request json: %v", err)
			}
			sort.Slice(metrics, func(i, j int) bool { return metrics[i].Name < metrics[j].Name })

			return map[string]interface{}{
				"common":  common,
				"metrics": metrics,
			}, nil
		},

		GetConfig: func(ctx interface{}) ([]string, error) {
			// This supplies the adapter config and some operator config for instances and rules
			// Metric template config is required and automatically added by the test framework
			s := ctx.(*Server)
			return []string{
				string(adapterCfg),
				fmt.Sprintf(string(operatorCfg), s.listener.Addr().String()),
			}, nil
		},

		// Want is the expected serialized json for the Result struct.
		// Result.AdapterState is what the callback function GetState returns.
		// Returns is going be full of empty values as this is a Report call not a Check call
		Want: `
        {
         "AdapterState": {
          "common": {
           "attributes": {
            "cluster.name": "hotdog-stand",
            "foodHandlerPermit": "revoked"
           }
          },
          "metrics": [
           {
            "attributes": {
             "connection.duration": 457200000,
             "connection.mtls": true,
             "connection.securityPolicy": "mutual_tls",
             "destination.app": "unknown",
             "destination.principal": "unknown",
             "destination.service": "unknown",
             "destination.service.name": "unknown",
             "destination.service.namespace": "unknown",
             "destination.version": "unknown",
             "destination.workload": "cat",
             "destination.workload.namespace": "unknown",
             "reporter": "destination",
             "request.protocol": "http",
             "request.time": 1582230020000,
             "response.code": 418,
             "response.flags": "-",
             "service.name": "cat",
             "source.app": "grocery",
             "source.ip": "1.2.3.4",
             "source.principal": "unknown",
             "source.version": "unknown",
             "source.workload": "unknown",
             "source.workload.namespace": "unknown"
            },
            "name": "istio.request.bytes",
            "type": "gauge",
            "value": 0
           },
           {
            "attributes": {
             "connection.duration": 457200000,
             "connection.mtls": true,
             "connection.securityPolicy": "mutual_tls",
             "destination.app": "unknown",
             "destination.principal": "unknown",
             "destination.service": "unknown",
             "destination.service.name": "unknown",
             "destination.service.namespace": "unknown",
             "destination.version": "unknown",
             "destination.workload": "cat",
             "destination.workload.namespace": "unknown",
             "reporter": "destination",
             "request.protocol": "http",
             "request.time": 1582230020000,
             "response.code": 418,
             "response.flags": "-",
             "service.name": "cat",
             "source.app": "grocery",
             "source.ip": "1.2.3.4",
             "source.principal": "unknown",
             "source.version": "unknown",
             "source.workload": "unknown",
             "source.workload.namespace": "unknown"
            },
            "name": "istio.request.duration.seconds",
            "type": "summary",
            "value": {
             "count": 1,
             "max": 0,
             "min": 0,
             "sum": 0
            }
           },
           {
            "attributes": {
             "connection.duration": 457200000,
             "connection.mtls": true,
             "connection.securityPolicy": "mutual_tls",
             "destination.app": "unknown",
             "destination.principal": "unknown",
             "destination.service": "unknown",
             "destination.service.name": "unknown",
             "destination.service.namespace": "unknown",
             "destination.version": "unknown",
             "destination.workload": "cat",
             "destination.workload.namespace": "unknown",
             "reporter": "destination",
             "request.protocol": "http",
             "request.time": 1582230020000,
             "response.code": 418,
             "response.flags": "-",
             "service.name": "cat",
             "source.app": "grocery",
             "source.ip": "1.2.3.4",
             "source.principal": "unknown",
             "source.version": "unknown",
             "source.workload": "unknown",
             "source.workload.namespace": "unknown"
            },
            "name": "istio.request.total",
            "type": "count",
            "value": 1
           },
           {
            "attributes": {
             "connection.duration": 457200000,
             "connection.mtls": true,
             "connection.securityPolicy": "mutual_tls",
             "destination.app": "unknown",
             "destination.principal": "unknown",
             "destination.service": "unknown",
             "destination.service.name": "unknown",
             "destination.service.namespace": "unknown",
             "destination.version": "unknown",
             "destination.workload": "cat",
             "destination.workload.namespace": "unknown",
             "reporter": "destination",
             "request.protocol": "http",
             "request.time": 1582230020000,
             "response.code": 418,
             "response.flags": "-",
             "service.name": "cat",
             "source.app": "grocery",
             "source.ip": "1.2.3.4",
             "source.principal": "unknown",
             "source.version": "unknown",
             "source.workload": "unknown",
             "source.workload.namespace": "unknown"
            },
            "name": "istio.response.bytes",
            "type": "gauge",
            "value": 0
           }
          ]
         },
         "Returns": [
          {
           "Check": {
            "Status": {},
            "ValidDuration": 0,
            "ValidUseCount": 0,
            "RouteDirective": null
           },
           "Quota": null,
           "Error": {}
          }
         ]
        }`,
	}

	integration.RunTest(t, nil, scenario)
}

// MockTransport caches decompressed request bodies
type MockTransport struct {
	requests [][]byte
}

func (c *MockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// telemetry sdk gzip compresses json payloads
	gz, err := gzip.NewReader(r.Body)
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	contents, err := ioutil.ReadAll(gz)
	if err != nil {
		return nil, err
	}

	if !json.Valid(contents) {
		return nil, errors.New("error validating request body json")
	}
	c.requests = append(c.requests, contents)

	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(&bytes.Buffer{}),
	}, nil
}

func newMockTransport() *MockTransport {
	return &MockTransport{
		requests: make([][]byte, 0),
	}
}

func mockHarvester(mt *MockTransport, common map[string]interface{}) (*instrumentation.MetricAggregator, *telemetry.Harvester) {
	agg := instrumentation.NewMetricAggregator()
	return agg, telemetry.NewHarvester(
		telemetry.ConfigAPIKey("8675309"),
		telemetry.ConfigCommonAttributes(common),
		telemetry.ConfigHarvestPeriod(time.Duration(5*time.Second)),
		telemetry.ConfigBasicErrorLogger(os.Stderr),
		telemetry.ConfigBasicDebugLogger(os.Stderr),
		agg.BeforeHarvest,
		func(cfg *telemetry.Config) {
			cfg.MetricsURLOverride = "localhost"
			cfg.SpansURLOverride = "localhost"
			cfg.Client.Transport = mt
		},
	)
}

type CommonAttributes struct {
	timestamp  interface{}       `json:"-"`
	interval   interface{}       `json:"-"`
	Attributes map[string]string `json:"attributes"`
}

type Metric struct {
	Name       string                 `json:"name"`
	Typo       string                 `json:"type"`
	Value      interface{}            `json:"value"`
	timestamp  interface{}            `json:"-"`
	Attributes map[string]interface{} `json:"attributes"`
}

func readTelemetryRequestJson(data []byte) (CommonAttributes, []Metric, error) {
	// Expected format:
	// [{
	//   "common": CommonAttributes{},
	//   "metrics": [Metric{}],
	// }]
	var objs []*json.RawMessage
	if err := json.Unmarshal(data, &objs); err != nil {
		return CommonAttributes{}, nil, err
	}

	var objmap map[string]*json.RawMessage
	if err := json.Unmarshal(*objs[0], &objmap); err != nil {
		return CommonAttributes{}, nil, err
	}

	var common CommonAttributes
	if err := json.Unmarshal(*objmap["common"], &common); err != nil {
		return CommonAttributes{}, nil, err
	}

	var metrics []Metric
	if err := json.Unmarshal(*objmap["metrics"], &metrics); err != nil {
		return CommonAttributes{}, nil, err
	}

	return common, metrics, nil
}
