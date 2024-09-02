package syscallextra

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"iter"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// Importable version of WrapEINTR
func WrapEINTR(wrapped func() error) error {
	for {
		if err := wrapped(); !errors.Is(err, unix.EINTR) {
			return err
		}
	}
}

// Calls statx, which gives more details than os.Stat().Sys()
func Statx(path string) (st unix.Statx_t, retErr error) {
	if err := WrapEINTR(func() error { return unix.Statx(-1, path, 0, 0, &st) }); err != nil {
		retErr = &os.PathError{Op: "statx", Path: path, Err: err}
	}
	return
}

// get 6:0 or similar for the given mount path
func DeviceMajorMinorFromMountPath(path string) (string, error) {
	if st, err := Statx(path); err != nil {
		return fmt.Sprintf("(%s)", err), err
	} else {
		return fmt.Sprintf("%d:%d", st.Dev_major, st.Dev_minor), nil
	}
}

// Opens path fd
func PathFD(ctx context.Context, path string) iter.Seq2[int, error] {
	return func(yield func(int, error) bool) {
		f, err := os.OpenFile(path, unix.O_PATH, 0)
		if err != nil {
			yield(-1, fmt.Errorf("%w (O_PATH)", err))
			return
		}
		defer f.Close()
		defer runtime.KeepAlive(f)
		yield(int(f.Fd()), nil)
	}
}

// inspect a given open FD return the proc value for it
func DescribeOpenFD(fd int) (string, error) {
	procPath := fmt.Sprintf("/proc/self/fd/%d", fd)
	return os.Readlink(procPath)
}

// Test helper that requires that FD be closed or open with a different description
func RequiredFDClosed(t *testing.T, fd int, shouldNotBe string) {
	t.Helper()
	fdstr, err := DescribeOpenFD(fd)
	if errors.Is(err, fs.ErrNotExist) {
		// Its closed, so return
		return
	}
	require.NoError(t, err)
	// If its open, it might be reopened, the description should be different
	require.NotEqual(t, shouldNotBe, fdstr)

}
