package lostandfound

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Truly misc things that have no home of their own
// Random stuff gets added here, then maybe moved somewhere more specific once
// there is a better place for it

// Small helper to make it easy to reset a value
// You can do something like this:
//
//	enableSomething := False
//	...
//	defer SetReset(&enableSomething, true)()
//	( enableSomething will be True until the function exits)
func SetReset[T any](variable *T, newValue T) func() {
	originalValue := *variable
	*variable = newValue
	return func() {
		// TODO: Add a runtime check to ensure that func gets called?
		*variable = originalValue
	}
}

// Removes all pending elements from a channel,
// returns when channel is exhausted
func DainChannel[T any](c <-chan T) {
	for {
		select {
		case <-c:
		default:
			return
		}
	}
}

// return a new slice created from applying f to each item in s
func MapApply[S any, R any](s []S, f func(S) R) []R {
	r := []R{}
	for _, ss := range s {
		r = append(r, f(ss))
	}
	return r
}

// return a new slice created from applying f to each item in s
// f is called with the index into the array
func MapApplyIdx[S any, R any](s []S, f func(int, S) R) []R {
	r := []R{}
	for idx, ss := range s {
		r = append(r, f(idx, ss))
	}
	return r
}

type MultipleError struct {
	Op   string
	Errs []error
}

func (e MultipleError) Error() string {
	s := strings.Builder{}
	fmt.Fprintf(&s, "OP: %s Errors:", e.Op)
	for i, err := range e.Errs {
		if i != 0 {
			fmt.Fprintf(&s, ", ")
		}
		fmt.Fprintf(&s, "(%s)", err)
	}
	return s.String()
}

func (e MultipleError) Unwrap() []error {
	return e.Errs
}

// This isn't really optimized, I suspect that perhaps using maps would be faster,
// but at that point, we'd probably want to just implement or employ a "Set" type
func SliceSubtract(a []string, b []string) []string {
	ret := []string{}
	for _, ai := range a {
		if !slices.Contains(b, ai) {
			ret = append(ret, ai)
		}
	}
	return ret
}

func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func RequireFileContent(t *testing.T, path, content string) {
	t.Helper()
	c, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, content, strings.TrimSpace(string(c)))
}
