package loggingsugar

import (
	"context"
	"log/slog"
	"slices"

	"gitlab.com/croepha/common-utils/errordump"
	"gitlab.com/croepha/common-utils/logging"
	"gitlab.com/croepha/common-utils/loggingctx"
)

/*

This package provides some alternative syntax sugar for using slog



*/

type L struct {
	parent    *L
	attr      slog.Attr
	wrapDepth int
}

// Increments the wrap depth, which causes more frames to be skipped when resolving the source lines (only works when directly logging)
func (l L) Wrap() L {
	l.wrapDepth++
	return l
}

// Disables source code location resolution (only works when directly logging)
func (l L) NoSource() L {
	l.wrapDepth = logging.WrapDepth__DisablePC
	return l
}

func (l *L) resolveAttrs() []slog.Attr {
	attrs := make([]slog.Attr, 0, 10)
	for l != nil {
		attrs = append(attrs, l.attr)
		l = l.parent
	}
	slices.Reverse(attrs)
	return attrs
}

// Add attribute
func (l L) A(name string, value any) L {
	return L{parent: &l, attr: slog.Any(name, value), wrapDepth: l.wrapDepth}
}

// Adds an error attribute to be deeply described
func (l L) Err(err error) L {
	return L{parent: &l, attr: errordump.NewSlog("error", err), wrapDepth: l.wrapDepth}
}

// Addes attribute
func (l L) Attr(attr slog.Attr) L {
	return L{parent: &l, attr: attr, wrapDepth: l.wrapDepth}
}

// Gives new handler with given attributes
func (l L) Handler(ctx context.Context) slog.Handler {
	return loggingctx.Handler(ctx).WithAttrs(l.resolveAttrs())
}

// Gives a new ctx that has a handler with these attributes
func (l L) Context(ctx context.Context) context.Context {
	return loggingctx.Context(ctx, l.Handler(ctx))
}

// handles a new log record
func (l L) LogFromWrapper(ctx context.Context, additionalWrapDepth int, level slog.Level, msg string) {
	logging.LogAttrs(ctx, loggingctx.Handler(ctx), l.wrapDepth+additionalWrapDepth+1, level, l.resolveAttrs(), msg)
}

// handles a new log record
func (l L) Log(ctx context.Context, level slog.Level, msg string) {
	l.LogFromWrapper(ctx, 1, level, msg)
}

// handles a new Debug log record
func (l L) Debug(ctx context.Context, msg string) {
	l.LogFromWrapper(ctx, 1, slog.LevelDebug, msg)
}

// handles a new Info log record
func (l L) Info(ctx context.Context, msg string) {
	l.LogFromWrapper(ctx, 1, slog.LevelInfo, msg)
}

// handles a new Warn log record
func (l L) Warn(ctx context.Context, msg string) {
	l.LogFromWrapper(ctx, 1, slog.LevelWarn, msg)
}

// handles a new Error log record
func (l L) Error(ctx context.Context, msg string) {
	l.LogFromWrapper(ctx, 1, slog.LevelError, msg)
}
