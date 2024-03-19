package helper

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestSetGlobalLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		level string
	}{
		{"DebugLevel", "debug"},
		{"InfoLevel", "info"},
		{"ErrorLevel", "error"},
		{"FatalLevel", "fatal"},
		{"DefaultLevel", "error"}, // Testing default case
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetGlobalLogLevel(tt.level)

			// Check if the global log level is set correctly
			expectedLevel := zerolog.GlobalLevel()

			if expectedLevel.String() != tt.level {
				t.Errorf("expected global log level to be %s, got %s", tt.level, expectedLevel)
			}
		})
	}
}
