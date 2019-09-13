# Development Guide

This document is intended to aid contributors of this project get up and running.

## Developing Locally

To begin developing locally, you will need the following things installed, configured, and verified to be working.

*   [git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)
*   [Docker](https://docs.docker.com/install/)
*   [Go](https://golang.org/doc/install#install) (>=1.13)
*   [protoc](https://github.com/protocolbuffers/protobuf)
    *   For example on MacOS:
        ```shell
        brew install protobuf
        ```
    *   Or change the `protoc` variable in `bin/mixer_codegen.sh` to run a Docker container instead
*   A local copy of [Mixer](https://github.com/istio/istio/tree/master/mixer)
    *   For example:
        ```shell
        mkdir -p $GOPATH/src/istio.io/ && \
          git clone https://github.com/istio/istio $GOPATH/src/istio.io/istio
        ```
*   A New Relic Insights Insert API Key.

Once you have these things in place, use the following steps to get a Mixer server running and test your build of the adapter with it.

### 1) Setup Istio Development Environment Variables

```shell
export MIXER_REPO=$GOPATH/src/istio.io/istio/mixer
export ISTIO=$GOPATH/src/istio.io
```

### 2) Build the Mixer Server, Client, and Toolset

```shell
pushd $ISTIO/istio && make mixs mixc mixgen && popd
```

### 3) Make a development copy of `config/` into `testdata/`

```shell
mkdir testdata
cp sample_operator_cfg.yaml \
  config/newrelic.yaml \
  $MIXER_REPO/testdata/config/attributes.yaml \
  $MIXER_REPO/template/metric/template.yaml \
  $MIXER_REPO/template/tracespan/tracespan.yaml \
  testdata
```

### 4) Start the newrelic-istio-adapter

Be sure to replace `"$NEW_RELIC_API_KEY"` with your valid New Relic Insights Insert API Key.

Another option is to export an environment variable (`NEW_RELIC_API_KEY`) prior to starting the `newrelic-istio-adapter`.

```shell
go run cmd/main.go \
  --cluster-name "Local Testing" \
  --debug \
  "$NEW_RELIC_API_KEY"
```

### 5) Start the Mixer Server

The location of the `mixs` binary depends on what OS you are on.

*   Linux:
    ```shell
    export MIXS="$GOPATH/out/linux_amd64/release/mixs"
    ```
*   MacOS:
    ```shell
    export MIXS="$GOPATH/out/darwin_amd64/release/mixs"
    ```

Start the Istio Mixer server with `configStoreURL` pointed to the `testdata` directory previously created.

```shell
$MIXS server --configStoreURL=fs://$(pwd)/testdata
```

Mixer should begin printing logs to `STDOUT`.
It is important to stop here and troubleshoot any errors reported by the Mixer server related to connecting to the locally running adapter.

### 6) Send Test Instances

The location of the `mixc` binary depends on what OS you are on.

*   Linux
    ```shell
    export MIXC="$GOPATH/out/linux_amd64/release/mixc"
    ```
*   MacOS
    ```shell
    export MIXC="$GOPATH/out/darwin_amd64/release/mixc"
    ```

Send test events to the Mixer server using the Mixer client.

```shell
$MIXC report \
  --string_attributes "context.protocol=http,destination.principal=service-account-bar,destination.service.host=bar.test.svc.cluster.local,destination.service.name=bar,destination.service.namespace=test,destination.workload.name=bar,destination.workload.namespace=test,source.principal=service-account-foo,source.service.host=foo.test.svc.cluster.local,source.service.name=foo,source.service.namespace=test,source.workload.name=foo,source.workload.namespace=test" \
  --stringmap_attributes "request.headers=x-forward-proto:https;source:foo,destination.labels=app:bar;version:v1,source.labels=app:foo" \
  --int64_attributes response.duration=2003,response.size=1024,response.code=200 \
  --timestamp_attributes "request.time="2017-07-04T00:01:10Z,response.time="2017-07-04T00:01:11Z" \
  --bytes_attributes source.ip=c0:0:0:2
```

The `newrelic-istio-adapter` should output activity to `STDOUT` signifying that metrics have been captured and sent to New Relic.

## Release Process

1.  Bump the project version.
    ```shell
    ./bin/bump_version.sh <new-version>
    ```
    The `<new-version>` needs to be an appropriately incremented [semver](https://semver.org/) format version.
    *   This means that any backwards-incompatible API changes being included in the release require the major version to be incremented.
    *   Any backwards-compatible changes that add functionality means the minor version must be incremented.
    *   If the release only contains bug fixes and patches, just the patch version will need to be incremented.
2.  Review the [CHANGELOG](./CHANGELOG.md) and ensure all relevant changes included in the release are captured there.
3.  Commit changes and open an appropriately titled (i.e. `Release X.X.X`) PR in GitHub.
    Have another developer approve your PR and then merge your changes to `master`.
4.  Create a GitHub release for the merged changes.
    Be sure to tag the commit of the release with the `<new-version>` value (do not include a `v` prefix) and have the release description body include all the changes from the [CHANGELOG](./CHANGELOG.md).
