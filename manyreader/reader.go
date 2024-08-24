package joboutput

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"
)

// Implements io.reader
type reader struct {
	position     int // Current position in output, in bytes
	sharedSource *writer
	ctx          context.Context
	logger       *slog.Logger
}

// Creates a new reader.  start can be used to skip forward to a place in
// output stream.  The reader read operations will be cancelled with the
// given ctx here.
// This function is safe to be called by multiple goroutines.  However the
// returned Reader, should not be used by multiple goroutines at the same time
func (s *writer) NewReader(ctx context.Context, start int) io.Reader {
	logger := s.logger.With("reader_id", nextReaderId.Add(1)-1)
	logger.Debug("created")

	return &reader{
		position:     start,
		sharedSource: s,
		ctx:          ctx,
		logger:       logger,
	}
}

var nextReaderId atomic.Uint64

// Implements io.Reader
func (c *reader) Read(out []byte) (int, error) {
	debugWaits := 0

	for {
		c.sharedSource.lock.Lock()
		signal := c.sharedSource.notifySignal

		// Make a copy of the slice
		// TODO: Could probably move the copy outside of the lock, would explore that
		// later when focusing on performance and scale
		copyLen := copy(out, c.sharedSource.output[c.position:])

		sourceEnd := len(c.sharedSource.output)

		c.position += copyLen

		// TODO: Would consider using a closure with defer unlock.  I think its impossible for
		// the above code to panic but it would be unfortunate to tie up the whole output
		// for all the readers because of a panic
		c.sharedSource.lock.Unlock()

		c.logger.DebugContext(c.ctx, "joboutput Read",
			"position_before_read", c.position-copyLen,
			"position_increment", copyLen,
			"position_limit", sourceEnd,
			"position_increment_limit", len(out),
			"writer_closed", signal == nil,
			"wait_iteration_count", debugWaits,
		)

		// if we have reached the end and source is closed, return EOF
		if c.position >= sourceEnd && signal == nil {
			return copyLen, io.EOF
		}

		// If we have some new output then return
		if copyLen > 0 {
			return copyLen, nil
		}

		debugWaits++

		// If we do not have any new output, then lets wait for the signal or cancel
		select {
		case <-c.ctx.Done():
			return 0, fmt.Errorf("manyreader Read cancelled:%w", c.ctx.Err())
		case <-signal:
			// signal is closed, which means new output or never anymore output
			// so we continue loop to try again
		}

	}

}
