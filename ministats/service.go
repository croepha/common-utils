package ministats

import (
	"sync"
	"time"
)

/*
This is a very simple stats collection and printing system that doesn't
require any external infrastructure

Features:
 - Prints values from counters.
 - Only prints values when there is a change
 - Doesn't print more than once every second (or given cooldown time)
*/

type Service struct {
	groups     map[string]*group
	mut        sync.RWMutex
	wakerCh    chan<- struct{}
	shutdownCh <-chan struct{}

	// May have data race if you change any of the following
	// while printer is running
	PrintOutput   OutputHandler
	PrintCooldown time.Duration
}

// Start the background printing routine
// output defines how stats are printed, it defaults to using SlogStatOutput
// NOP if already started
func (s *Service) Start() {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.wakerCh == nil {
		if s.PrintOutput == nil {
			s.PrintOutput = NewDefaultStatOutput()
		}
		if s.PrintCooldown == 0 {
			s.PrintCooldown = time.Second
		}

		wakeCh := make(chan struct{}, 1)
		shutCh := make(chan struct{})
		s.wakerCh = wakeCh
		s.shutdownCh = shutCh
		go s.printInBackground(wakeCh, shutCh)
		s.wake()
	}
}

// Stops the printing service
// NOP if it isn't running
func (s *Service) Stop() {
	s.mut.Lock()
	wakeCh := s.wakerCh
	shutCh := s.shutdownCh
	s.wakerCh = nil
	s.shutdownCh = nil
	s.mut.Unlock()

	if wakeCh != nil {
		close(wakeCh)
		<-shutCh
	}
}

func (s *Service) printInBackground(waker <-chan struct{}, shutCh chan<- struct{}) {
	defer close(shutCh)
	values := map[string]uint64{}
	for range waker { // Wait for wakeups in a loop, there is no backlog

		if s.loadAllStats(values) {
			s.PrintOutput(values)
		}

		// TODO: This could cause a momentary stall on calls to Stop(), to fix it
		// we would need to add some kind of special cooldown bypass channel
		time.Sleep(s.PrintCooldown)
	}
}

func (s *Service) wake() {
	select { // Wake up the logger if needed
	case s.wakerCh <- struct{}{}:
	default:
	}
}
