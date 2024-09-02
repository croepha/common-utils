package ministats

import "log/slog"

// Output implementation that outputs to a given slog.Logger
type SlogOutput struct {
	Logger *slog.Logger
}

func (o *SlogOutput) Write(values map[string]uint64) {
	for k, v := range values {
		o.Logger.Info("stat", "name", k, "value", v)
	}
}

type OutputHandler func(map[string]uint64)

func NewDefaultStatOutput() OutputHandler {
	o := SlogOutput{
		Logger: slog.Default().With(
			"package", "stats",
		),
	}
	return o.Write
}
