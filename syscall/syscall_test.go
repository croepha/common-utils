package syscall_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/syscall"
	"golang.org/x/sys/unix"
)

func TestWrapEINTROtherError(t *testing.T) {
	expectedErr := errors.New("expected")
	require.ErrorIs(t, syscall.WrapEINTR(func() error {
		return expectedErr
	}), expectedErr)
}

func TestWrapEINTRNilError(t *testing.T) {
	require.NoError(t, syscall.WrapEINTR(func() error {
		return nil
	}))
}

func TestWrapEINTRRetryOtherError(t *testing.T) {
	expectedErr := errors.New("expected")
	counter := 5
	require.ErrorIs(t, syscall.WrapEINTR(func() error {
		if counter > 0 {
			counter--
			return unix.EINTR
		}
		return expectedErr
	}), expectedErr)
	require.Equal(t, 0, counter)
}

func TestWrapEINTRRetryNilError(t *testing.T) {
	counter := 5
	require.ErrorIs(t, syscall.WrapEINTR(func() error {
		if counter > 0 {
			counter--
			return unix.EINTR
		}
		return nil
	}), nil)
	require.Equal(t, 0, counter)
}
