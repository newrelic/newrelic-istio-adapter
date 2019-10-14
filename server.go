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

// nolint:lll
//go:generate ./bin/mixer_codegen.sh -a ./config/config.proto -x "-s=false -n newrelic -t metric -t tracespan"

package newrelic

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/newrelic/newrelic-istio-adapter/config"
	"github.com/newrelic/newrelic-istio-adapter/log"
	nrmetric "github.com/newrelic/newrelic-istio-adapter/metric"
	"github.com/newrelic/newrelic-istio-adapter/trace"
	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	adptModel "istio.io/api/mixer/adapter/model/v1beta1"
	"istio.io/istio/mixer/template/metric"
	"istio.io/istio/mixer/template/tracespan"
)

// Server listens for Mixer metrics and sends them to New Relic.
type Server struct {
	listener     net.Listener
	healthServer *health.Server
	server       *grpc.Server
	shutdown     chan error

	rawcfg      []byte
	builderLock sync.RWMutex

	handler *Handler

	harvester *telemetry.Harvester
}

// Compile time assertion Server implement what it is expected to.
var (
	_ metric.HandleMetricServiceServer       = &Server{}
	_ tracespan.HandleTraceSpanServiceServer = &Server{}
)

// NewServer creates a new Mixer metric handling server.
func NewServer(addr string, h *telemetry.Harvester, grpcOpt ...grpc.ServerOption) (*Server, error) {
	s := &Server{
		rawcfg:       []byte{0xff, 0xff},
		healthServer: health.NewServer(),
		server:       grpc.NewServer(grpcOpt...),
		harvester:    h,
	}

	var err error
	if s.listener, err = net.Listen("tcp", addr); err != nil {
		return nil, fmt.Errorf("unable to listen on %q: %v", addr, err)
	}
	log.Infof("listening on %q", s.listener.Addr().String())

	metric.RegisterHandleMetricServiceServer(s.server, s)
	tracespan.RegisterHandleTraceSpanServiceServer(s.server, s)
	if _, err = s.getHandler(nil); err != nil {
		return nil, err
	}

	s.healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s.server, s.healthServer)

	return s, nil
}

// getHandler returns the handler for rawcfg if it already exists otherwise it builds it.
func (s *Server) getHandler(rawcfg []byte) (*Handler, error) {
	s.builderLock.RLock()
	if bytes.Equal(rawcfg, s.rawcfg) {
		h := s.handler
		s.builderLock.RUnlock()
		return h, nil
	}
	s.builderLock.RUnlock()

	// build a handler for the rawcfg and establish session.
	cfg := &config.Params{}
	if err := cfg.Unmarshal(rawcfg); err != nil {
		return nil, err
	}

	s.builderLock.Lock()
	defer s.builderLock.Unlock()

	// check again if someone else beat you to this.
	if bytes.Equal(rawcfg, s.rawcfg) {
		return s.handler, nil
	}

	mh, err := nrmetric.BuildHandler(cfg, s.harvester)
	if err != nil {
		return nil, err
	}

	th, err := trace.BuildHandler(cfg, s.harvester)
	if err != nil {
		return nil, err
	}

	s.rawcfg = rawcfg
	s.handler = &Handler{m: mh, t: th}

	return s.handler, nil
}

// Run starts the Server.
func (s *Server) Run() {
	s.shutdown = make(chan error, 1)
	go func() {
		err := s.server.Serve(s.listener)

		// notify closer we're done
		s.shutdown <- err
	}()
}

// Wait waits for Server to stop.
func (s *Server) Wait() error {
	if s.shutdown == nil {
		return fmt.Errorf("server not running")
	}

	err := <-s.shutdown
	s.shutdown = nil
	return err
}

// Close gracefully shuts down Server.
func (s *Server) Close() error {
	var results error
	if s.shutdown != nil {
		s.healthServer.Shutdown()
		s.server.GracefulStop()
		if err := s.Wait(); err != nil {
			results = err
		}
	}

	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			if results != nil {
				results = fmt.Errorf("%v: %w", err, results)
			} else {
				results = err
			}
		}
	}

	return results
}

// HandleTraceSpan implements tracespan.HandleMetricServiceServer.
func (s *Server) HandleTraceSpan(ctx context.Context, r *tracespan.HandleTraceSpanRequest) (*adptModel.ReportResult, error) {
	h, err := s.getHandler(r.AdapterConfig.Value)
	if err != nil {
		return nil, err
	}

	if err = h.HandleTraceSpan(ctx, r.Instances); err != nil {
		return nil, fmt.Errorf("failed to handle tracespan: %v", err)
	}

	return &adptModel.ReportResult{}, nil
}

// HandleMetric implements metric.HandleMetricServiceServer.
func (s *Server) HandleMetric(ctx context.Context, r *metric.HandleMetricRequest) (*adptModel.ReportResult, error) {
	h, err := s.getHandler(r.AdapterConfig.Value)
	if err != nil {
		return nil, err
	}

	if err = h.HandleMetric(ctx, r.Instances); err != nil {
		return nil, fmt.Errorf("failed to handle metric: %v", err)
	}

	return &adptModel.ReportResult{}, nil
}
