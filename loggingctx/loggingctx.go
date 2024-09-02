package loggingctx

import (
	"context"
	"log/slog"
	"slices"

	"gitlab.com/croepha/common-utils/logging"
)

/*

Some tools to make it easy to ctx based logging
A Slog handler can be added to and retrieved from the givent context

Debug/Info/Warn/Error are provided as convenience functions that use
Handler from ctx

*/

type contextKeyType string

var contextKey contextKeyType = "std-logging"

// Gets Handler from context (or use the default if one isn't set)
func Handler(ctx context.Context) slog.Handler {
	v, _ := ctx.Value(contextKey).(slog.Handler)
	if v == nil {
		v = slog.Default().Handler()
	}
	return v
}

// Creates a new context with the given handler added to it
func Context(parent context.Context, handler slog.Handler) context.Context {
	return context.WithValue(parent, contextKey, handler)
}

// Log a Debug record using handler from context
func Debug(ctx context.Context, msg string, args ...any) {
	logging.Log(ctx, Handler(ctx), 1, slog.LevelDebug, args, msg)
}

// Log an Info record using handler from context
func Info(ctx context.Context, msg string, args ...any) {
	logging.Log(ctx, Handler(ctx), 1, slog.LevelInfo, args, msg)
}

// Log a Warn record using handler from context
func Warn(ctx context.Context, msg string, args ...any) {
	logging.Log(ctx, Handler(ctx), 1, slog.LevelWarn, args, msg)
}

// Log an Error record using handler from context
func Error(ctx context.Context, msg string, args ...any) {
	logging.Log(ctx, Handler(ctx), 1, slog.LevelError, args, msg)
}

var originalDefaultSlogHandler slog.Handler

func init() {
	originalDefaultSlogHandler = slog.Default().Handler()
}

// Create a new handler instance
// This handler simply uses the supplied ctx to check for a Handler and uses it
// This handler can be used directly or installed as slog.Default()
func NewHandler() *ctxHandler {
	return &ctxHandler{}
}

type ctxHandler struct {
	attrs  []slog.Attr
	groups []string
}

func (h *ctxHandler) handler(ctx context.Context) slog.Handler {
	real := Handler(ctx)
	if _, ok := real.(*ctxHandler); ok {
		real = originalDefaultSlogHandler
		slog.New(real).ErrorContext(ctx, "recursive use of handler abaited")
	}
	return Handler(ctx)
}

func (h *ctxHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.handler(ctx).Enabled(ctx, l)
}

func (h *ctxHandler) Handle(ctx context.Context, r slog.Record) error {
	real := h.handler(ctx)
	if len(h.attrs) > 0 {
		real = real.WithAttrs(h.attrs)
	}
	for _, g := range h.groups {
		real = real.WithGroup(g)
	}
	return real.Handle(ctx, r)
}

func (h *ctxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	r := *h
	r.attrs = slices.Concat(r.attrs, attrs)
	return &r
}

func (h *ctxHandler) WithGroup(name string) slog.Handler {
	r := *h
	r.groups = slices.Concat(r.groups, []string{name})
	return &r
}
