package errordump_test

import (
	"context"
	"fmt"
	"log/slog"
	"syscall"
	"testing"

	"gitlab.com/croepha/common-utils/errordump"
	"gitlab.com/croepha/common-utils/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCDetailer(err error) errordump.Details {
	if err, _ := err.(interface{ GRPCStatus() *status.Status }); err != nil {
		return err.GRPCStatus().Proto()
	}
	return nil
}

func Test(t *testing.T) {

	ctx := context.Background()
	th := logging.NewTestHandler(t)
	handler := th.H
	logger := slog.New(handler)

	errordump.SetGlobalDetailer(
		errordump.NewUnwrappingDetailer(
			errordump.ReflectionDetailer(
				errordump.ChainDetailers(
					GRPCDetailer,
					errordump.RawDetailer,
				),
			),
		),
	)

	e := fmt.Errorf("wrap test: %w %w %w",
		status.Errorf(codes.Aborted, "asdf"),
		fmt.Errorf("test"),
		fmt.Errorf("wrap test: %w", syscall.E2BIG),
	)

	type o = map[string]any
	logger.ErrorContext(ctx, "error", errordump.NewSlog("error", e))
	th.RequireLine(true, slog.LevelError, "error", "error", o{
		"Error": o{
			"NextDetails":          o{},
			"ReflectedName":        "wrapErrors",
			"ReflectedPackagePath": "fmt",
			"String":               "wrap test: rpc error: code = Aborted desc = asdf test wrap test: argument list too long",
		}, "WrappedErrors": []o{
			{
				"NextDetails": o{
					"code": 10, "message": "asdf",
				},
				"ReflectedName":        "Error",
				"ReflectedPackagePath": "google.golang.org/grpc/internal/status",
				"String":               "rpc error: code = Aborted desc = asdf",
			},
			{
				"NextDetails":          o{},
				"ReflectedName":        "errorString",
				"ReflectedPackagePath": "errors",
				"String":               "test",
			},
			{
				"Error": o{
					"NextDetails":          o{},
					"ReflectedName":        "wrapError",
					"ReflectedPackagePath": "fmt",
					"String":               "wrap test: argument list too long"},
				"WrappedError": o{
					"NextDetails":          7,
					"ReflectedName":        "Errno",
					"ReflectedPackagePath": "syscall",
					"String":               "argument list too long",
				},
			},
		},
	})
}
