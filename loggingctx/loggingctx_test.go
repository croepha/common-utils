package loggingctx_test

import (
	"context"
	"log/slog"
	"testing"

	"gitlab.com/croepha/common-utils/logging"
	"gitlab.com/croepha/common-utils/loggingctx"
)

func TestInfo(t *testing.T) {
	ctx := context.Background()
	th := logging.NewTestHandler(t)
	handler := th.H

	ctx = loggingctx.Context(ctx, handler)

	loggingctx.Info(ctx, "info test", "attr0", "foo")
	th.RequireLine(true, slog.LevelInfo, "info test", "attr0", "foo")
}

func TestHandler(t *testing.T) {
	ctx := context.Background()
	th := logging.NewTestHandler(t)
	handler := th.H

	ctx = loggingctx.Context(ctx, handler)

	l := slog.New(loggingctx.NewHandler())

	l.InfoContext(ctx, "info test", "attr0", "foo")
	th.RequireLine(true, slog.LevelInfo, "info test", "attr0", "foo")

	l.With("with0", "with0").InfoContext(ctx, "info test", "attr0", "foo")
	th.RequireLine(true, slog.LevelInfo, "info test", "with0", "with0", "attr0", "foo")

	l.WithGroup("withGroup0").InfoContext(ctx, "info test", "attr0", "foo")
	th.RequireLine(true, slog.LevelInfo, "info test", "withGroup0", map[string]any{"attr0": "foo"})

}
