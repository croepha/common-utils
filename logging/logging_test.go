package logging_test

import (
	"context"
	"log/slog"
	"testing"

	"gitlab.com/croepha/common-utils/logging"
)

func TestLog(t *testing.T) {
	ctx := context.Background()
	th := logging.NewTestHandler(t)
	handler := th.H

	logging.Log(ctx, handler, 0, slog.LevelInfo, []any{"attr1", 10}, "test Log")
	th.RequireLine(true, slog.LevelInfo, "test Log", "attr1", 10)
}

func TestLogAttr(t *testing.T) {
	ctx := context.Background()
	th := logging.NewTestHandler(t)
	handler := th.H

	logging.LogAttrs(ctx, handler, 0, slog.LevelInfo, []slog.Attr{slog.Any("attr1", 10)}, "test Log")
	th.RequireLine(true, slog.LevelInfo, "test Log", "attr1", 10)

}

// type TestStruct struct {
// 	M0 string
// }

func TestLogAttr2(t *testing.T) {
	ctx := context.Background()
	th := logging.NewTestHandler(t)
	handler := th.H

	logging.LogAttrs(ctx, handler, 0, slog.LevelInfo, []slog.Attr{slog.Any("attr1", 10)}, "test Log")
	th.RequireLine(true, slog.LevelInfo, "test Log", "attr1", 10)

}
