package telemetry

import (
	"bytes"
	"strings"
	"testing"
)

func TestConfigAPIKey(t *testing.T) {
	apikey := "apikey"
	h := NewHarvester(ConfigAPIKey(apikey))
	if h.config.APIKey != apikey {
		t.Error("config func does not set APIKey correctly")
	}
}

func TestConfigHarvestPeriod(t *testing.T) {
	h := NewHarvester(ConfigHarvestPeriod(0))
	if 0 != h.config.HarvestPeriod {
		t.Error("config func does not set harvest period correctly")
	}
}

func TestConfigBasicErrorLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	h := NewHarvester(ConfigBasicErrorLogger(buf))

	buf.Reset()
	h.config.logError(map[string]interface{}{"zip": "zap"})
	if log := buf.String(); !strings.Contains(log, "{\"zip\":\"zap\"}") {
		t.Error("message not logged correctly", log)
	}

	buf.Reset()
	h.config.logError(map[string]interface{}{"zip": func() {}})
	if log := buf.String(); !strings.Contains(log, "json: unsupported type: func()") {
		t.Error("message not logged correctly", log)
	}
}

func TestConfigBasicDebugLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	h := NewHarvester(ConfigBasicDebugLogger(buf))

	buf.Reset()
	h.config.logDebug(map[string]interface{}{"zip": "zap"})
	if log := buf.String(); !strings.Contains(log, "{\"zip\":\"zap\"}") {
		t.Error("message not logged correctly", log)
	}

	buf.Reset()
	h.config.logDebug(map[string]interface{}{"zip": func() {}})
	if log := buf.String(); !strings.Contains(log, "json: unsupported type: func()") {
		t.Error("message not logged correctly", log)
	}
}

func TestConfigAuditLogger(t *testing.T) {
	h := NewHarvester(configTesting)
	if enabled := h.config.auditLogEnabled(); enabled {
		t.Error("audit logging should not be enabled", enabled)
	}
	// This should not panic.
	h.config.logAudit(map[string]interface{}{"zip": "zap"})

	buf := new(bytes.Buffer)
	h = NewHarvester(configTesting, ConfigBasicAuditLogger(buf))
	if enabled := h.config.auditLogEnabled(); !enabled {
		t.Error("audit logging should be enabled", enabled)
	}
	h.config.logAudit(map[string]interface{}{"zip": "zap"})
	if lg := buf.String(); !strings.Contains(lg, `{"zip":"zap"}`) {
		t.Error("audit message not logged correctly", lg)
	}
}
