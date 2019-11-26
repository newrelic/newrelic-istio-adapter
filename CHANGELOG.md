# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Unreleased

### Added

*   Added GID and UID 2000 to Dockerfile. The `newrelic-istio-adapter` container now runs as non-root user. 

## 2.0.0

### Added

*   Travis CI configuration to validate and build this project.
*   A `--log-level` command flag to specify the user desired minimum logging level.
*   A `log` package to unify the adapter logging function.

### Changed

*   Switched to using the upstream [New Relic Go telemetry SDK](https://github.com/newrelic/newrelic-telemetry-sdk-go) instead of the internal `nrsdk` package.
*   Unified the adapter logging. Logs should now have a unified format and be configurable globaly to set the logging level.
*   The `metric.BuildHandler` and `trace.BuildHandler` functions no longer take an Istio adapter `Logger` interface as an argument.
*   The helm-charts now have a `logLevel` value to specify the adapter logging level during the deploy.

### Removed

*   The `--debug` command flag is now replaced by setting the `--log-level` flag to `debug`.

## 1.0.0

### Added

*   Initial `newrelic-istio-adapter` application code.
*   Documentation, user guides, and project metadata.
*   Build configuration files.
