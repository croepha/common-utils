package joboutput

import (
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/logging"
)

func TestCorrectness(t *testing.T) {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jow := NewWriter()

	// Ensure we confrom to the interface
	var w io.WriteCloser = jow

	write := func(s string) {
		n, err := w.Write([]byte(s))
		require.NoError(t, err)
		require.Equal(t, len(s), n)
	}

	// Do a read in the background, and then through the returned function assert the results
	// Most of the time we just want to wait on and assert the results right away, but there
	// are a couple places where we would block and we want to do something that would
	// break the block by doing a context cancel, writing some data, or closing the writer
	read := func(c io.Reader, limit int) func(expectedErr error, expectedStr string) {
		var readErr error
		var readStr string
		join := make(chan struct{})
		go func() {
			defer close(join)
			o := make([]byte, limit)
			n, err := c.Read(o)
			readStr = string(o[:n])
			readErr = err
		}()
		return func(expectedErr error, expectedStr string) {
			t.Helper()
			select {
			case <-join:
				require.ErrorIs(t, readErr, expectedErr)
				require.Equal(t, expectedStr, readStr)
			case <-time.After(5 * time.Second):
				t.Fatal("time out")
			}
		}
	}

	// Buffer multiple adds and read once
	c0 := jow.NewReader(ctx, 0)
	write("aaaaa")
	write("bbbbb")
	read(c0, 10000)(nil, "aaaaabbbbb")

	// Add some more and read it
	write("ccccc")
	read(c0, 10000)(nil, "ccccc")

	// New cursor, it should read all the content
	c1 := jow.NewReader(ctx, 0)
	read(c1, 10000)(nil, "aaaaabbbbbccccc")

	// Another new cursor, but we skip forward 3 bytes, and read more in 3 byte chunks
	c2 := jow.NewReader(ctx, 3)
	read(c2, 3)(nil, "aab")
	read(c2, 3)(nil, "bbb")
	read(c2, 3)(nil, "bcc")
	read(c2, 3)(nil, "ccc")

	// Add some more after reading, and also read in cursors in reverse
	// order to show that there is no coupling between them, and there are no
	// deadlocks
	w0 := read(c0, 10000)
	w1 := read(c1, 10000)
	w2 := read(c2, 10000)
	time.Sleep(1 * time.Millisecond)
	write("ddddd")
	w2(nil, "ddddd")
	w1(nil, "ddddd")
	w0(nil, "ddddd")

	{
		// Setup a new ctx so we can cancel it specifically
		ctx, cancel := context.WithCancel(ctx)
		c2 := jow.NewReader(ctx, 0)
		read(c2, 10000)(nil, "aaaaabbbbbcccccddddd")

		w := read(c2, 10000) // This should block
		cancel()
		w(context.Canceled, "")
	}

	// Finalize the output, reads shouldn't block
	err := w.Close()
	require.NoError(t, err)

	read(c0, 10000)(io.EOF, "")
	read(c1, 10000)(io.EOF, "")
	read(c2, 10000)(io.EOF, "")

	// Another new cursor, this should read
	// everything and never block
	c4 := jow.NewReader(ctx, 0)
	read(c4, 10000)(io.EOF, "aaaaabbbbbcccccddddd")
	read(c4, 10000)(io.EOF, "")

	//  Small reads after close
	r5 := jow.NewReader(ctx, 0)
	read(r5, 3)(nil, "aaa")
	read(r5, 3)(nil, "aab")
	read(r5, 3)(nil, "bbb")
	read(r5, 3)(nil, "bcc")
	read(r5, 3)(nil, "ccc")
	read(r5, 3)(nil, "ddd")
	read(r5, 3)(io.EOF, "dd")

}

func TestCloseClosed(t *testing.T) {
	jo := NewWriter()
	require.NoError(t, jo.Close())
	require.Nil(t, jo.notifySignal)
	require.ErrorIs(t, jo.Close(), os.ErrClosed)

}

func TestMain(m *testing.M) {
	logging.SlogStartup()
	m.Run()
}
