package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerLevel(t *testing.T) {
	cases := []struct {
		name     string
		level    Level
		expected Level
	}{
		{name: "debug", level: DebugLevel, expected: DebugLevel},
		{name: "info", level: InfoLevel, expected: InfoLevel},
		{name: "warn", level: WarnLevel, expected: WarnLevel},
		{name: "error", level: ErrorLevel, expected: ErrorLevel},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			log := New(Config{Level: tc.level, Output: &bytes.Buffer{}})
			assert.Equal(t, tc.expected, log.Level())
		})
	}
}
