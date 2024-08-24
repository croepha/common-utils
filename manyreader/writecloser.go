package joboutput

import (
	"log/slog"
	"os"
	"sync"
	"sync/atomic"
)

/*
This implements a single writer multiple concurrent reader pattern.
Every new reader reads the whole buffer.
Buffer is stored in memory
*/

// This implements the io.WriteCloser interface
type writer struct {
	// This lock should be acquired while operating on
	// anything in this struct
	// TODO: I think RWMutex may be good here, as we will likely
	// have one writer, and many readers, but we should wait and
	// see if this locking is actually a bottleneck
	lock sync.Mutex

	// TODO: If we needed persistence, storing log output in
	// a file or sqlite would be better, but using an in-memory
	// slice for now because it's easy
	// NOTE: Related to above, probably could have pluggable storage...
	output []byte

	// This is used to notify any waiters that we have new output
	notifySignal chan struct{}

	logger *slog.Logger
}

// Create a new Writer
func NewWriter() *writer {
	logger := slog.Default().With(
		"package", "manyreader", // To differentiate from any potential future `writers`
		"writer_id", nextWriterId.Add(1)-1,
	)
	logger.Debug("created")
	return &writer{
		notifySignal: make(chan struct{}),
		logger:       logger,
	}
}

var nextWriterId atomic.Uint64

// Implements io.Writer, add the new output to the in-memory buffer
// wakes up any waiters that are blocking
func (s *writer) Write(newOutput []byte) (int, error) {

	s.logger.Debug("write",
		"write_length", len(newOutput),
		"write_snippet", newOutput[:min(30, len(newOutput))],
	)

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.notifySignal == nil {
		return 0, os.ErrClosed
	}

	// TODO: Should profile this, may need to manually handle slice capacity
	// if the default growslice behavior isn't great for this workflow
	// Also maybe a copy would be better.  Leaving the simple code here until
	// there is a reason to do something different...

	// TODO: If this copy proves to be too expensive, then we could explore
	// reading directly into the end of the buffer. I think we can do that
	// without taking a lock for the read.  But we'd need to make sure that
	// wasn't a race
	s.output = append(s.output, newOutput...)
	// TODO: Look into implementing io.ReaderFrom to pair with the above strategy

	// TODO: This could be reworked so that when there are no readers, the
	// channel is nil, and the first Reader that blocked would
	// allocate and set the signal.  This could save a useless allocation for
	// the cases where there are no readers, which in-theory would have
	// performance implications.  I would like to do some benchmarks on the
	// current implementaton, with no readers, and see that compared to the
	// same setup with the channel allocation commented out.  This would be
	// and interesting baseline to see how impactful the allocation is anyway
	// Saving the perf explorations for later if there is time.
	close(s.notifySignal)                // Notify any blocking waiters
	s.notifySignal = make(chan struct{}) // Reset for new waiters

	return len(newOutput), nil
}

// Implements io.Closer, this signals to all the readers that there is no more data and they
// wont block
func (s *writer) Close() error {

	s.logger.Debug("close")

	s.lock.Lock()
	defer s.lock.Unlock()

	if s.notifySignal == nil {
		return os.ErrClosed
	}

	close(s.notifySignal)
	s.notifySignal = nil
	return nil
}
