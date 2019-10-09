package log

import (
	"encoding/json"

	"github.com/newrelic/newrelic-telemetry-sdk-go/telemetry"
	iLog "istio.io/pkg/log"
)

const (
	// AdapterScopeName is the default logging scope name for this adapter.
	AdapterScopeName       = "newrelic"
	defaultOutputLevel     = ErrorLevel
	defaultStackTraceLevel = NoneLevel
)

var adapterScope = iLog.RegisterScope(AdapterScopeName, "New Relic adapter logging messages.", 0)

func init() {
	adapterScope.SetOutputLevel(defaultOutputLevel.istioLevel())
	adapterScope.SetStackTraceLevel(defaultStackTraceLevel.istioLevel())

	// Configure the default Istio logger with our default options.
	// This ensures things like gRPC logging (which is overriden by the
	// Istio log package) is configured at the appropriate level.
	opts := iLog.DefaultOptions()
	opts.SetOutputLevel(iLog.DefaultScopeName, defaultOutputLevel.istioLevel())
	opts.SetStackTraceLevel(iLog.DefaultScopeName, defaultStackTraceLevel.istioLevel())
	_ = iLog.Configure(opts)
}

// Fatalf uses fmt.Sprintf to construct and log a message at fatal level.
func Fatalf(template string, args ...interface{}) {
	adapterScope.Fatalf(template, args...)
}

// Errorf uses fmt.Sprintf to construct and log a message at error level.
func Errorf(template string, args ...interface{}) {
	adapterScope.Errorf(template, args...)
}

// Warnf uses fmt.Sprintf to construct and log a message at warn level.
func Warnf(template string, args ...interface{}) {
	adapterScope.Warnf(template, args...)
}

// Infof uses fmt.Sprintf to construct and log a message at info level.
func Infof(template string, args ...interface{}) {
	adapterScope.Infof(template, args...)
}

// Debugf uses fmt.Sprintf to construct and log a message at debug level.
func Debugf(template string, args ...interface{}) {
	adapterScope.Debugf(template, args...)
}

// SetOutputLevel adjusts the output level associated with the adapter scope.
func SetOutputLevel(l Level) {
	adapterScope.SetOutputLevel(l.istioLevel())
}

// SetStackTraceLevel adjusts the stack tracing level associated with the adapter scope.
func SetStackTraceLevel(l Level) {
	adapterScope.SetStackTraceLevel(l.istioLevel())
}

// telemetryLogger returns a suitable telemetry.Harvester logging function
func telemetryLogger(logf func(string, ...interface{})) func(map[string]interface{}) {
	return func(fields map[string]interface{}) {
		if js, err := json.Marshal(fields); nil != err {
			logf("%s", err.Error())
		} else {
			logf("%s", string(js))
		}
	}
}

// HarvesterConfigFunc returns a configuration function for the telemetry Harvester initialization.
//
// There is not a one-to-one mapping of our/Istio logging levels to the telemetry package.
// The mapping made here is that error logging is at the ErrorLevel, debug logging is at
// the InfoLevel, and audit logging is at the DebugLevel.
func HarvesterConfigFunc() func(*telemetry.Config) {
	return func(cfg *telemetry.Config) {
		cfg.ErrorLogger = telemetryLogger(adapterScope.Errorf)
		cfg.DebugLogger = telemetryLogger(adapterScope.Infof)
		cfg.AuditLogger = telemetryLogger(adapterScope.Debugf)
	}
}
