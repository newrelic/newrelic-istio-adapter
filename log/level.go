package log

import (
	"fmt"

	iLog "istio.io/pkg/log"
)

// Level of logging severity
type Level int

const (
	// InvalidLevel represents an invalid logging level returned when parsing errors occur.
	InvalidLevel Level = iota

	// DebugLevel signifies all messages with debug level and above should be logged.
	DebugLevel

	// InfoLevel signifies all messages with info level and above should be logged.
	InfoLevel

	// WarnLevel signifies all messages with warn level and above should be logged.
	WarnLevel

	// ErrorLevel signifies all messages with fatal level and above should be logged.
	ErrorLevel

	// FatalLevel signifies only fatal level messages should be logged.
	FatalLevel

	// NoneLevel signifies no messages be logged.
	NoneLevel
)

var stringToLevel = map[string]Level{
	"debug": DebugLevel,
	"info":  InfoLevel,
	"warn":  WarnLevel,
	"error": ErrorLevel,
	"fatal": FatalLevel,
	"none":  NoneLevel,
}

var levelToString = map[Level]string{
	DebugLevel: "debug",
	InfoLevel:  "info",
	WarnLevel:  "warn",
	ErrorLevel: "error",
	FatalLevel: "fatal",
	NoneLevel:  "none",
}

var levelToIstioLevel = map[Level]iLog.Level{
	DebugLevel: iLog.DebugLevel,
	InfoLevel:  iLog.InfoLevel,
	WarnLevel:  iLog.WarnLevel,
	ErrorLevel: iLog.ErrorLevel,
	FatalLevel: iLog.FatalLevel,
	NoneLevel:  iLog.NoneLevel,
}

// ParseLevel interprets level name as a Level.
func ParseLevel(name string) (Level, error) {
	if s, ok := stringToLevel[name]; ok {
		return s, nil
	}
	return InvalidLevel, fmt.Errorf("invalid log level: %s", name)
}

// String returns Level string representation.
func (l Level) String() string {
	return levelToString[l]
}

func (l Level) istioLevel() iLog.Level {
	return levelToIstioLevel[l]
}

// Levels returns all valid level names.
func Levels() []string {
	l := make([]string, 0, len(stringToLevel))
	for k := range stringToLevel {
		l = append(l, k)
	}
	return l
}
