package syscallextra_test

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/syscallextra"
	"golang.org/x/sys/unix"
)

func TestWrapEINTROtherError(t *testing.T) {
	expectedErr := errors.New("expected")
	require.ErrorIs(t, syscallextra.WrapEINTR(func() error {
		return expectedErr
	}), expectedErr)
}

func TestWrapEINTRNilError(t *testing.T) {
	require.NoError(t, syscallextra.WrapEINTR(func() error {
		return nil
	}))
}

func TestWrapEINTRRetryOtherError(t *testing.T) {
	expectedErr := errors.New("expected")
	counter := 5
	require.ErrorIs(t, syscallextra.WrapEINTR(func() error {
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
	require.ErrorIs(t, syscallextra.WrapEINTR(func() error {
		if counter > 0 {
			counter--
			return unix.EINTR
		}
		return nil
	}), nil)
	require.Equal(t, 0, counter)
}

func TestRootMajorMinor(t *testing.T) {

	statOut, err := exec.Command("stat", "/", "-c%Hd:%Ld").Output()
	require.NoError(t, err)

	d, err := syscallextra.DeviceMajorMinorFromMountPath("/")
	require.NoError(t, err)
	require.Equal(t, strings.TrimSpace(string(statOut)), d)

}

func TestPathFD(t *testing.T) {

	path := t.TempDir()
	for fd, err := range syscallextra.PathFD(context.Background(), path) {
		require.NoError(t, err)
		check, err := syscallextra.DescribeOpenFD(fd)
		require.NoError(t, err)
		require.Equal(t, path, check)
	}
}
