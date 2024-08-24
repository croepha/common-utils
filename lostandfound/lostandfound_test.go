package lostandfound_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/lostandfound"
)

func TestSetReset(t *testing.T) {

	v := 3

	func() {
		defer lostandfound.SetReset(&v, 7)()
		require.Equal(t, 7, v)
	}()
	require.Equal(t, 3, v)

}
