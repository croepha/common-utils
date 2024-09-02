package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

/*
 General logging test helpers
*/

// This returns a struct that helps unit test logging wrappers
// the member H is a Handler that can be used as a slog.Handler
// See example usage in TestTestHandler
func NewTestHandler(t *testing.T) *testHandler {
	th := &testHandler{t: t}
	th.H = slog.NewJSONHandler(
		&th.buf, &slog.HandlerOptions{
			AddSource: true,
		},
	)
	return th
}

type testHandler struct {
	H   slog.Handler
	buf bytes.Buffer
	t   *testing.T
}

// Requires that a specific line was logged
// source specifies whether or not the source location information should be expected
// This should be called imediately after the log, and it should be done in the same function as you expect
// the source location to be reported
func (th *testHandler) RequireLine(source bool, expectedLevel slog.Level, expectedMsg string, expectedArgs ...any) {
	th.t.Helper()

	type jsonObject = map[string]any

	expectedBuf := bytes.Buffer{}
	slog.New(slog.NewJSONHandler(&expectedBuf, nil)).
		Log(context.Background(), expectedLevel, expectedMsg, expectedArgs...)
	expected := jsonObject{}
	require.NoError(th.t, json.Unmarshal(expectedBuf.Bytes(), &expected))
	delete(expected, "time")

	if source {
		_, sourcePath, sourceLine, ok := runtime.Caller(1)
		require.True(th.t, ok)
		expected["source"] = jsonObject{
			"file": sourcePath,
			"line": float64(sourceLine - 1),
		}
	}

	actual := jsonObject{}
	require.NoError(th.t, json.Unmarshal(th.buf.Bytes(), &actual))
	delete(actual, "time")
	if s := actual["source"]; s != nil {
		delete(s.(jsonObject), "function")
	}
	require.Equal(th.t, expected, actual)
	th.buf.Reset()
}
