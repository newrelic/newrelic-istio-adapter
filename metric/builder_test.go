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

package metric

import (
	"strings"
	"testing"

	adapterTest "istio.io/istio/mixer/pkg/adapter/test"

	"github.com/newrelic/newrelic-istio-adapter/config"
)

func TestBuildConfig(t *testing.T) {
	testCases := []struct {
		isValid        bool
		namespace      string
		metricName     string
		expectedErrMsg string
	}{
		{true, "", "request.bytes", ""},
		{true, "istio", "request.bytes", ""},
		{true, "istio.", "istio..request.bytes", ""},
		{true, "istio", "123", ""},
		{false, "istio", "", invalidNonEmptyErrMsg},
		{false, "", "", invalidNonEmptyErrMsg},
		// these test cases check that names are <= 255 UTF-16 code units; each code unit is 2 bytes
		{false, "", strings.Repeat("z", 256), invalidTooLongErrMsg},
		// U+61 a is encoded to 1 UTF-16 code unit, so this string is 255 runes and 255 code units
		{true, "", strings.Repeat("a", 254) + string('\U00000061'), ""},
		// U+13189 Egyptian Hieroglyph turtle is encoded to 2 UTF-16 code units, so this string is 255 runes and 256 code units
		{false, "", strings.Repeat("a", 254) + string('\U00013189'), invalidTooLongErrMsg},
	}

	for i, tc := range testCases {
		pmi := &config.Params_MetricInfo{
			Name: tc.metricName,
			Type: config.GAUGE,
		}

		params := &config.Params{
			Namespace: tc.namespace,
			Metrics:   map[string]*config.Params_MetricInfo{"example": pmi},
		}

		errString := ""
		_, err := buildConfig(params)

		if err != nil {
			errString = err.Error()
		}
		if tc.isValid && err != nil {
			t.Errorf("index %d: Did not expect error, got %s", i, errString)
		}
		if !tc.isValid && !strings.Contains(errString, tc.expectedErrMsg) {
			t.Errorf("index %d: Expected error to contain '%s', got %s", i, tc.expectedErrMsg, errString)
		}
	}
}

func TestBuildHandler(t *testing.T) {
	mixerMetricName := "requestsize.instance.istio-system"

	testCases := []struct {
		isValid      bool
		namespace    string
		metricName   string
		expectedName string
	}{
		{true, "", "request.bytes", "request.bytes"},
		{true, "istio", "request.bytes", "istio.request.bytes"},
		{true, "istio.", "request.bytes", "istio..request.bytes"},
		{true, "istio", "123", "istio.123"},
		{false, "istio", "", ""},
	}

	for _, tc := range testCases {
		pmi := &config.Params_MetricInfo{
			Name: tc.metricName,
			Type: config.GAUGE,
		}

		params := &config.Params{
			Namespace: tc.namespace,
			Metrics:   map[string]*config.Params_MetricInfo{mixerMetricName: pmi},
		}

		env := adapterTest.NewEnv(t)
		h, err := BuildHandler(params, nil, env)

		// Invalid name, but we saw expected error, so it's all good.
		if !tc.isValid && err != nil {
			continue
		}

		// Invalid name, but we didn't see expected error.
		if !tc.isValid && err == nil {
			t.Errorf("expected error for invalid metric '%s', but did not get one.", tc.metricName)
			continue
		}

		// Valid name, but we got an error building the handler.
		if tc.isValid && err != nil {
			t.Errorf("metric '%s' is valid, but got error building handler: '%v'", tc.metricName, err)
			continue
		}

		// Valid name, so make sure we see it in metrics map.
		if tc.isValid && err == nil {
			_, ok := h.metrics[mixerMetricName]
			if !ok {
				t.Errorf("metric '%s' is valid, but not found in metrics map: '%#v'.", tc.metricName, h.metrics)
			}
		}
	}
}
