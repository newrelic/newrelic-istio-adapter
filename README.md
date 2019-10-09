[![Go Report Card](https://goreportcard.com/badge/github.com/newrelic/newrelic-istio-adapter)](https://goreportcard.com/report/github.com/newrelic/newrelic-istio-adapter)
[![Build Status](https://travis-ci.com/newrelic/newrelic-istio-adapter.svg?branch=master)](https://travis-ci.com/newrelic/newrelic-istio-adapter)

# New Relic Istio Adapter

*An Istio Mixer adapter to send telemetry data to New Relic.*

For more information on how Istio Mixer telemetry is created and collected, please see this [Mixer Overview](https://istio.io/docs/reference/config/policy-and-telemetry/mixer-overview/). 

For more information about out-of-process Istio Mixer adapters, please see the [Mixer Out of Process Adapter Walkthrough](https://github.com/istio/istio/wiki/Mixer-Out-Of-Process-Adapter-Walkthrough)

## Quotas

**Metrics and Spans exported from this adapter to New Relic will be rate limited!**

Currently (2019-08-30) the following quotas apply to APM Professional accounts:

*   500,000 Metrics / minute
    *   250,000 unique Metric timeseries
    *   50 attributes per Metric
*   5,000 Spans / minute

You may request a quota increase for Metrics and/or Span by contacting your New Relic account representative.

## Quickstart

The `newrelic-istio-adapter` should be run alongside an installed/configured Istio Mixer server.

For Kubernetes installations, Helm deployment charts have been provided in the `helm-charts` directory.
These charts are intended to provide a simple installation and customization method for users.

See the [Helm installation docs](https://helm.sh/docs/using_helm/#install-helm) for installation/configuration of Helm.

### Prerequisites

*   A Kubernetes cluster
*   A working `kubectl` installation
*   A working `helm` installation
*   A Healthy Istio deployment
*   A New Relic Insights Insert API Key.

### Deploy Helm Template

The `newrelic-istio-adapter` should be deployed to an independent namespace.
This provides isolation and customizable access control.

The examples in this guide install the adapter to the `newrelic-istio-adapter` namespace.
This namespace is not managed by this installation process and will need to be created manually.
I.e.

```shell
kubectl create namespace newrelic-istio-adapter
```

Additionally, several components of the `newrelic-istio-adapter` are required to be deployed into the Istio namespace (i.e. `istio-system`).
Make sure you have privileges to deploy to this namespace.

Once you have ensured all of these things, generate Kubernetes manifests with Helm (be sure to replace `<your_new_relic_api_key>` with your New Relic Insights API key) and deploy the components using `kubectl`.

```shell
cd helm-charts
helm template . \
    -f values.yaml \
    --name newrelic-istio-adapter \
    --namespace newrelic-istio-adapter \
    --set authentication.apiKey=<your_new_relic_api_key> \
    > newrelic-istio-adapter.yaml
kubectl apply -f newrelic-istio-adapter.yaml
```

### Validate

Verify that the `newrelic-istio-adapter` deployment and pod are healthy within the `newrelic-istio-adapter` namespace

```
$ kubectl -n newrelic-istio-adapter get deploy newrelic-istio-adapter

NAME                     READY   UP-TO-DATE   AVAILABLE   AGE
newrelic-istio-adapter   1/1     1            1           10s

$ kubectl -n newrelic-istio-adapter get po -l app.kubernetes.io/name=newrelic-istio-adapter

NAME                                      READY   STATUS    RESTARTS   AGE
newrelic-istio-adapter-6d9c4f9b88-r5gn7   1/1     Running   1          8s
```

Verify that the `newrelic-istio-adapter` handler, rules, adapter, and instances are present within the `istio-system` namespace 

```
$ kubectl -n istio-system get handler -l app.kubernetes.io/name=newrelic-istio-adapter

NAME                     AGE
newrelic-istio-adapter   10s

$ kubectl -n istio-system get rules -l app.kubernetes.io/name=newrelic-istio-adapter

NAME                             AGE
newrelic-http-connection         10s
newrelic-tcp-connection          10s
newrelic-tcp-connection-closed   10s
newrelic-tcp-connection-open     10s

$ kubectl -n istio-system get adapter -l app.kubernetes.io/name=newrelic-istio-adapter

NAME       AGE
newrelic   10s

$ kubectl -n istio-system get instances -l app.kubernetes.io/name=newrelic-istio-adapter

NAME                          AGE
newrelic-bytes-received       10s
newrelic-bytes-sent           10s
newrelic-connections-closed   10s
newrelic-connections-opened   10s
newrelic-request-count        10s
newrelic-request-duration     10s
newrelic-request-size         10s
newrelic-response-size        10s
newrelic-span                 10s
```

You should start to see metrics sent to [Insights](https://insights.newrelic.com) a few minutes after the deployment.
As an example, this Insights query will display a timeseries graph of total Istio requests:

```
From Metric SELECT sum(istio.request.total) TIMESERIES
```

The logs of Mixer and the `newrelic-istio-adapter` in Kubernetes should show activity or errors.

```shell
kubectl -n newrelic-istio-adapter logs -l app.kubernetes.io/name=newrelic-istio-adapter
```

To get started visualizing your data try the [sample dashboard template](#new-relic-dashboard-template).

### Clean Up

If you want to remove the `newrelic-istio-adapter` you can do so by deleting the resources defined in the manifest you deployed.

```
kubectl delete -f newrelic-istio-adapter.yaml
```

## Distributed Tracing

The `newrelic-istio-adapter` is able to send [trace spans from services within the Istio service mesh](https://istio.io/docs/tasks/telemetry/distributed-tracing/overview/) to New Relic.
This functionality is disabled by default, but it can be enabled by adding the following `telemetry.rules` value when deploying the `newrelic-istio-adapter` Helm [chart](./helm-charts/README.md#configuration).

```
...
newrelic-tracing:
  match: (context.protocol == "http" || context.protocol == "grpc") && destination.workload.name != "istio-telemetry" && destination.workload.name != "istio-pilot" && ((request.headers["x-b3-sampled"] | "0") == "1")
  instances:
    - newrelic-span
```

Adding this rule means that Mixer will send the adapter all `HTTP`/`gRPC` spans for services that propagate appropriate Zipkin (B3) headers in their requests.

Note that the match condition for this rule configures Mixer to only send spans that have been sampled (i.e. `x-b3-sampled: 1`).
It is up the services themselves to appropriately sample traces.

This sampling is important to keep in mind when enabling this functionality.
Without sampling you can quickly exceed the [quota](#quotas) associated with your account for the number of spans-per-minute you are allowed.
Additionally, the cost of sending spans to New Relic needs to be understood **before** you enable this.

## New Relic Dashboard Template

A [dashboard template](sample_newrelic_dashboard.json) is provided to chart some Istio metrics the default configuration produces. The template is designed to be imported with the [Insights Dashboard API](https://docs.newrelic.com/docs/insights/insights-api/manage-dashboards/insights-dashboard-api) and can be created straight from the [API Explorer](https://rpm.newrelic.com/api/explore/dashboards/create).

The sample dashboard can be filtered by `cluster.name`, `destination.service.name`, and `source.app`.

## Versioning

This project follows [semver](http://semver.org/).

See the [CHANGELOG](./CHANGELOG.md) for a detailed description of changes between versions.
