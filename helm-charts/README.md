# New Relic Istio Adapter Helm Chart

This [Helm](https://helm.sh/) chart enables you to deploy the `newrelic-istio-adapter` to your Kubernetes cluster.

## Configuration

This chart is designed to be configured to support your environment.
The following table lists the configurable parameters of the `newrelic-istio-adapter` chart and their default values.

| Parameter                           | Description                                                                                                                                                             | Default                                                     |
|-------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------|
| `image.repository`                  | Repository for container image.                                                                                                                                         | `newrelic/newrelic-istio-adapter`|
| `image.tag`                         | Image tag.                                                                                                                                                              | `latest`                                                    |
| `image.pullPolicy`                  | Image pull policy.                                                                                                                                                      | `IfNotPresent`                                              |
| `nameOverride`                      | Override the Chart name. Also used as the label `app.kubernetes.io/name` value for all resources.                                                                       | `""`                                                        |
| `fullnameOverride`                  | Override the naming of `newrelic-istio-adapter` namespace resources. Enables multiple `newrelic-istio-adapter` versions to be deployed simultaneously.                  | `""`                                                        |
| `istioNamespace`                    | Namespace of the Istio control plane resources.                                                                                                                         | `istio-system`                                              |
| `clusterName`                       | Name used by `newrelic-istio-adapter` as a unique Kubernetes cluster identifier for metrics sent to New Relic.                                                          | `istio-cluster`                                             |
| `logLevel`                          | Logging verbosity level for the `newrelic-istio-adapter`. Allowed values are: `debug`, `info`, `warn`, `error`, `fatal`, or `none`.                                     | `error`                                   |
| `authentication.manageSecret`       | Create a Kubernetes Secret to manage `authentication.apiKey` securely. Setting this to `false` mean you will manually handle secrets.                                   | `true`                                                      |
| `authentication.apiKey`             | New Relic API Key, deployed as a Kubernetes Secret. **Required if `authentication.manageSecret` is `true`.**                                                            | `""`                                                        |
| `authentication.secretNameOverride` | Name of the Kubernetes Secret providing your API key (stored in the `NEW_RELIC_API_KEY` field). **If set, `authentication.manageSecret` must be `false`.**              | `""`                                                        |
| `service.type`                      | The `newrelic-istio-adapter` Kubernetes Service type.                                                                                                                   | `ClusterIP`                                                 |
| `service.port`                      | The `newrelic-istio-adapter` Kubernetes Service port.                                                                                                                   | 80                                                          |
| `replicaCount`                      | Kubernetes Deployment relica count definition.                                                                                                                          | `1`                                                         |
| `resources`                         | Kubernetes Pod resource requests & limits resource definition.                                                                                                          | `{}`                                                        |
| `nodeSelector`                      | Kubernetes Deployment nodeSelector definition.                                                                                                                          | `{}`                                                        |
| `tolerations`                       | Kubernetes Deployment tolerations definition.                                                                                                                           | `[]`                                                        |
| `affinity`                          | Kubernetes Deployment affinity definition.                                                                                                                              | `{}`                                                        |
| `proxy.http`                        | Proxy server address to route HTTP traffic through.                                                                                                                     | No value set                                                |
| `proxy.https`                       | Proxy server address to route HTTPS traffic through.                                                                                                                    | No value set                                                |
| `proxy.none`                        | HTTP(S) endpoints to not route through configured proxies.                                                                                                              | No value set                                                |

| `telemetry.namespace`               | Prefixed name for all metrics sent from `newrelic-istio-adapter` to New Relic.                                                                                          | `istio`                                                     |
| `telemetry.attributes`              | **Advanced** Envoy to Mixer attribute mapping. See the [Istio Attribute Vocabulary](https://istio.io/docs/reference/config/policy-and-telemetry/attribute-vocabulary/). | *See [values.yaml](values.yaml)*                            |
| `telemetry.traces`                  | **Advanced** Istio [Tracespan](https://istio.io/docs/reference/config/policy-and-telemetry/templates/tracespan/) mapping.                                               | *See [values.yaml](values.yaml)*                            |
| `telemetry.metrics`                 | **Advanced** Istio [Metric](https://istio.io/docs/reference/config/policy-and-telemetry/templates/metric/) mapping.                                                     | *See [values.yaml](values.yaml)*                            |
| `telemetry.rules`                   | **Advanced** Istio [Policy and Telemetry Rules](https://istio.io/docs/reference/config/policy-and-telemetry/istio.policy.v1beta1/)                                      | *See [values.yaml](values.yaml)*                            |


Specify each parameter using the `--set key=value[,key=value]` argument to `helm template`.

Alternatively, a YAML file that specifies the values for the above parameters can be provided while installing the chart.
> **Tip**: You can use the default [values.yaml](values.yaml)
