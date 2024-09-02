package logging_test

import (
	"log/slog"
	"testing"

	"gitlab.com/croepha/common-utils/logging"
)

func TestTestHandler(t *testing.T) {

	th := logging.NewTestHandler(t)

	slog.New(th.H).Info("test message", "attr0", "foo", "attr1", "bar")
	th.RequireLine(true, slog.LevelInfo, "test message", "attr0", "foo", "attr1", "bar")

}
