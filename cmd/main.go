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

package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	newrelic "github.com/newrelic/newrelic-istio-adapter"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/instrumentation"
	"github.com/newrelic/newrelic-istio-adapter/internal/nrsdk/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/alecthomas/kingpin.v2"
)

// Version is the semver set during build with an ldflag arg.
// E.g. go build -ldflags "-X main.Version=0.1.0" ...
var Version = "undefined"

var (
	portPtr          = kingpin.Flag("port", "port gRPC server listens on").Default("55912").OverrideDefaultFromEnvar("NEW_RELIC_PORT").Short('p').Int32()
	clusterNamePtr   = kingpin.Flag("cluster-name", "Name of cluster where metrics come from").OverrideDefaultFromEnvar("NEW_RELIC_CLUSTER_NAME").String()
	debugPtr         = kingpin.Flag("debug", "enable debug logging").OverrideDefaultFromEnvar("NEW_RELIC_DEBUG").Short('d').Bool()
	harvestPeriodPtr = kingpin.Flag("harvest-period", "rate data is reported to New Relic").Default("5s").OverrideDefaultFromEnvar("NEW_RELIC_HARVEST_PERIOD").Duration()
	metricsHostPtr   = kingpin.Flag("metrics-host", "Endpoint to send metrics (used for debugging)").OverrideDefaultFromEnvar("NEW_RELIC_METRICS_HOST").String()
	spansHostPtr     = kingpin.Flag("spans-host", "Endpoint to send spans (used for debugging)").OverrideDefaultFromEnvar("NEW_RELIC_SPANS_HOST").String()
	mtlsCertPtr      = kingpin.Flag("cert", "mTLS certificate for gRPC server").OverrideDefaultFromEnvar("NEW_RELIC_MTLS_CERT").ExistingFile()
	mtlsKeyPtr       = kingpin.Flag("key", "mTLS key for gRPC server").OverrideDefaultFromEnvar("NEW_RELIC_MTLS_KEY").ExistingFile()
	mtlsCAPtr        = kingpin.Flag("ca", "mTLS CA certificate for gRPC server").OverrideDefaultFromEnvar("NEW_RELIC_MTLS_CA").ExistingFile()
	apiKeyPtr        = kingpin.Arg("api-key", "New Relic API key").Envar("NEW_RELIC_API_KEY").Required().String()
)

func getServerTLSOption(cert, key, ca string) (grpc.ServerOption, error) {
	certificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load key cert pair: %v", err)
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{certificate}}

	if ca != "" {
		certPool := x509.NewCertPool()
		bs, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("failed to read client ca cert %q: %v", ca, err)
		}

		if ok := certPool.AppendCertsFromPEM(bs); !ok {
			return nil, fmt.Errorf("failed to append client certs")
		}

		tlsConfig.ClientCAs = certPool
	}

	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert

	return grpc.Creds(credentials.NewTLS(tlsConfig)), nil
}

func main() {
	kingpin.Version(Version)
	kingpin.Parse()

	var commonAttrs map[string]interface{}
	if clusterNamePtr != nil && *clusterNamePtr != "" {
		commonAttrs = map[string]interface{}{
			"cluster.name": *clusterNamePtr,
		}
	}

	debugLogFile := ioutil.Discard
	if *debugPtr {
		debugLogFile = os.Stderr
	}

	agg := instrumentation.NewMetricAggregator()
	h := telemetry.NewHarvester(
		telemetry.ConfigAPIKey(*apiKeyPtr),
		telemetry.ConfigCommonAttributes(commonAttrs),
		telemetry.ConfigBasicErrorLogger(os.Stderr),
		telemetry.ConfigHarvestPeriod(*harvestPeriodPtr),
		telemetry.ConfigBasicDebugLogger(debugLogFile),
		agg.BeforeHarvest,
		func(cfg *telemetry.Config) {
			cfg.MetricsURLOverride = *metricsHostPtr
			cfg.SpansURLOverride = *spansHostPtr
		},
	)

	address := fmt.Sprintf(":%d", *portPtr)

	var err error
	var s *newrelic.Server
	if *mtlsCertPtr != "" && *mtlsKeyPtr != "" {
		so, err := getServerTLSOption(*mtlsCertPtr, *mtlsKeyPtr, *mtlsCAPtr)
		if err != nil {
			fmt.Printf("Unable to configure gRPC server TLS: %v\n", err)
			os.Exit(-1)
		}
		s, err = newrelic.NewServer(address, agg, h, so)
	} else {
		s, err = newrelic.NewServer(address, agg, h)
	}
	if err != nil {
		fmt.Printf("Unable to start server: %v\n", err)
		os.Exit(-1)
	}

	// Termination handler.
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-term:
			fmt.Println("Received SIGTERM, exiting gracefully...")
			if err := s.Close(); err != nil {
				fmt.Printf("%v\n", err)
			}
		}
	}()

	s.Run()
	if err := s.Wait(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
