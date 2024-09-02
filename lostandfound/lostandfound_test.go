package lostandfound_test

import (
	"strings"
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

func TestMapApply(t *testing.T) {
	require.Equal(t,
		[]string{"AAA", "BBB", "CCC"},
		lostandfound.MapApply(
			[]string{"aaa", "bbb", "ccc"},
			strings.ToUpper,
		),
	)

}

func TestSliceSubtract(t *testing.T) {
	require.Equal(t,
		strings.Split("AAA BBB CCC", " "),
		lostandfound.SliceSubtract(
			strings.Split("AAA BBB CCC DDD EEE", " "),
			strings.Split("DDD EEE FFF GGG HHH", " "),
		),
	)
}
