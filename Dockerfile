# Copyright 2019 New Relic Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#############################################################################
# STEP 1: build static binary
#############################################################################
FROM golang:1.13-stretch as builder

WORKDIR ${GOPATH}/src/github.com/newrelic/newrelic-istio-adapter
COPY . .

ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
RUN make build

#############################################################################
# STEP 2: create a small image
#############################################################################
FROM alpine:3.10
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /go/bin/newrelic-istio-adapter /go/bin/newrelic-istio-adapter
COPY THIRD_PARTY_NOTICES.md ./THIRD_PARTY_NOTICES
COPY LICENSE.md ./LICENSE

# gRPC health check utility
ARG GRPC_HEALTH_PROBE_VERSION=v0.2.0
RUN wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 \
    && chmod +x /bin/grpc_health_probe

# Create group and user, root privilege not needed
RUN addgroup -g 2000 newrelic && \
    adduser -S -D -H -u 2000 -G newrelic newrelic
USER 2000

EXPOSE 55912
ENTRYPOINT ["/go/bin/newrelic-istio-adapter"]
