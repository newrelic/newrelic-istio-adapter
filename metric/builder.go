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

package metric

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/newrelic/newrelic-istio-adapter/config"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"istio.io/istio/mixer/pkg/adapter"
)

const (
	invalidNonEmptyErrMsg = "must be non-empty"
	invalidTooLongErrMsg  = "must be <= 255 16-bit code units"
	maxCodeUnitsCount     = 255
)

type metricConfig struct {
	namespace string
	metrics   map[string]info
}

// BuildHandler returns a metric Handler with valid configuration.
func BuildHandler(params *config.Params, h *telemetry.Harvester, env adapter.Env) (*Handler, error) {
	cfg, err := buildConfig(params)
	if err != nil {
		return nil, err
	}
	env.Logger().Infof("Built metrics: %#v", cfg.metrics)

	handler := &Handler{
		logger:  env.Logger(),
		agg:     h.MetricAggregator(),
		metrics: cfg.metrics,
	}
	return handler, nil
}

// buildConfig returns a valid metricConfig. Most importantly, it iterates through
// the metrics from config.Params and validates them. If any of the metrics are
// invalid, it returns nil instead.
func buildConfig(params *config.Params) (cfg *metricConfig, errs *adapter.ConfigErrors) {
	namespace := params.GetNamespace()
	handlerMetrics := make(map[string]info, len(params.GetMetrics()))
	cfg = &metricConfig{
		namespace: namespace,
		metrics:   handlerMetrics,
	}

	for instName, mInfo := range params.GetMetrics() {
		if mInfo.GetType() == config.UNSPECIFIED {
			errs = errs.Appendf("MetricInfo.type", "type is invalid: UNSPECIFIED")
			cfg = nil
		}

		metricName := mInfo.GetName()
		if metricName == "" {
			errs = errs.Appendf("MetricInfo.name", "must be non-empty")
			cfg = nil
			continue
		}

		fullName := buildName(namespace, metricName)
		err := validateName(fullName)
		if err != nil {
			if namespace == "" {
				errs = errs.Appendf("MetricInfo.name", "name is invalid: %v", err)
			} else {
				errs = errs.Appendf("MetricInfo.name", "namespace + name is invalid: %v", err)
			}
			cfg = nil
			continue
		}

		// Add valid metric name, so long as we haven't seen any validation errors yet.
		if cfg != nil {
			cfg.metrics[instName] = info{
				name:  fullName,
				mtype: mInfo.GetType(),
			}
		}
	}
	return cfg, errs
}

func countUtf16CodeUnits(s string) int {
	codePoints := []rune(s)
	codeUnits := utf16.Encode(codePoints)
	return len(codeUnits)
}

func validateName(name string) error {
	// Must be non-empty.
	if name == "" {
		return fmt.Errorf(invalidNonEmptyErrMsg)
	}
	// Must be less than or equal to 255 16-bit code units (UTF-16).
	if countUtf16CodeUnits(name) > maxCodeUnitsCount {
		return fmt.Errorf("%q %s", name, invalidTooLongErrMsg)
	}
	return nil
}

// buildName creates the full name by combining namespace and MetricInfo.name.
func buildName(namespace, metricName string) string {
	name := metricName
	if namespace != "" {
		name = strings.Join([]string{namespace, metricName}, ".")
	}
	return name
}
