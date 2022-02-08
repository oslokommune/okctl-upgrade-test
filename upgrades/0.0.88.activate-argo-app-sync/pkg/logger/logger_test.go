package logger

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:funlen
func TestLevels(t *testing.T) {
	testCases := []struct {
		name         string
		withLevel    Level
		withInfo     string
		withDebug    string
		withError    string
		expectStdout string
		expectStderr string
	}{
		{
			name:      "Should write nothing when given nothing",
			withLevel: Info,
		},
		{
			name:         "Should only write info when only given info",
			withLevel:    Info,
			withInfo:     "debug",
			expectStdout: "debug\n",
		},
		{
			name:         "Should only write debug when only given debug",
			withLevel:    Debug,
			withDebug:    "debug",
			expectStdout: "[DEBUG] debug\n",
		},
		{
			name:         "Should write debug and info when given debug and info with debug level",
			withLevel:    Debug,
			withDebug:    "debug",
			withInfo:     "info",
			expectStdout: "info\n[DEBUG] debug\n",
		},
		{
			name:         "Should only write info when debug and info is given with info level",
			withLevel:    Info,
			withInfo:     "info",
			withDebug:    "debug",
			expectStdout: "info\n",
		},
		{
			name:         "Should only write error when everything is given and error level",
			withLevel:    Error,
			withInfo:     "info",
			withDebug:    "debug",
			withError:    "error",
			expectStderr: "[ERROR] error\n",
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			stdout := bytes.Buffer{}
			stderr := bytes.Buffer{}

			log := NewWithWriters(tc.withLevel, &stdout, &stderr)

			if tc.withInfo != "" {
				log.Info(tc.withInfo)
			}

			if tc.withDebug != "" {
				log.Debug(tc.withDebug)
			}

			if tc.withError != "" {
				log.Error(tc.withError)
			}

			assert.Equal(t, tc.expectStdout, stdout.String())
			assert.Equal(t, tc.expectStderr, stderr.String())
		})
	}
}
