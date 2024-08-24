package syscall

import (
	"errors"

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
