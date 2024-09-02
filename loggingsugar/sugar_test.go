package loggingsugar_test

import (
	"context"
	"log/slog"
	"testing"

	"gitlab.com/croepha/common-utils/logging"
	"gitlab.com/croepha/common-utils/loggingctx"
	"gitlab.com/croepha/common-utils/loggingsugar"
)

var l loggingsugar.L

func TestSugar(t *testing.T) {

	ctx := context.Background()

	th := logging.NewTestHandler(t)

	handler := th.H

	ctx = loggingctx.Context(ctx, handler)
	l.A("attr1", 10).A("attr2", 20).Info(ctx, "message")
	th.RequireLine(true, slog.LevelInfo, "message", "attr1", 10, "attr2", 20)

	func() {
		l.Wrap().A("attr1", 10).A("attr2", 20).Info(ctx, "message")
	}()
	th.RequireLine(true, slog.LevelInfo, "message", "attr1", 10, "attr2", 20)

	l.NoSource().A("attr1", 10).A("attr2", 20).Info(ctx, "message")
	th.RequireLine(false, slog.LevelInfo, "message", "attr1", 10, "attr2", 20)

}

func (h *NullHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *NullHandler) Handle(_ context.Context, _ slog.Record) error {
	return nil
}

func (h *NullHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *NullHandler) WithGroup(name string) slog.Handler {
	return h
}

type NullHandler struct {
}

func BenchmarkSugar(b *testing.B) {
	handler := &NullHandler{}
	ctx := context.Background()
	ctx = loggingctx.Context(ctx, handler)
	l := loggingsugar.L{}
	for i := range b.N {
		l.A("bench_i", i).Info(ctx, "test line")
	}
}

func BenchmarkBaselineLogger(b *testing.B) {
	handler := &NullHandler{}

	ctx := loggingctx.Context(context.Background(), handler)

	logger := slog.New(handler)
	for i := range b.N {
		logger.InfoContext(ctx, "test line", "bench_i", i)
	}
}

func BenchmarkBaselineAttrs(b *testing.B) {
	handler := &NullHandler{}

	ctx := loggingctx.Context(context.Background(), handler)

	logger := slog.New(handler)
	for i := range b.N {
		logger.LogAttrs(ctx, slog.LevelInfo, "test line", slog.Int("bench_i", i))
	}
}
