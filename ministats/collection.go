package ministats

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type group struct {
	clct   *Service
	locked bool // Used locally, to see if we need to unlock on afterChange

	mut sync.Mutex // Used to control when the collector reads our values
	// mut Protects the following members:
	vals    map[string]*atomic.Uint64
	removed bool // Signal to collector that this should be removed

}

/*
 NOTE/TODO:

 using g.mut to exclude from collection seems to work fine... but maybe it wouldn't at larger scales... perhaps
 intead we should try removing from the map temporarily? Also, could do some things with queing like put all new
 items in a new queue... but TBH this code as it is now fast enough for my needs..

*/

// Optionally, this can be called before making changes to stats to ensure that
// updates to the group aren't read partially.  The stats printer will only print
// the state _before_ the changes, or _after_ but not in-between.
func (g *group) BeforeChange() {
	g.mut.Lock()
	g.locked = true
}

// Owners of tracked stats should call this after a change
// This ensures that the printer wakes some time after this is called
func (g *group) AfterChange() {
	if g.locked {
		g.mut.Unlock()
		g.locked = false
	}
	g.clct.wake()
}

// Removes the group from the collector
func (g *group) Remove() {
	// Failsafe check: Just making sure final stats are printed, and also that we are unlocked
	g.AfterChange()

	g.mut.Lock()
	defer g.mut.Unlock()

}

// Creates a new group of counters to monitor
// The given counters can be identified by the key of the map
// This returns a handle to the group to control varous
// things... Most notablly, AfterChange() which should be called after updating
// the counters
func (s *Service) AddNamed(name string, counters map[string]*atomic.Uint64) *group {
	s.mut.Lock()
	defer s.mut.Unlock()

	if s.groups == nil {
		s.groups = map[string]*group{}
	}

	if _, ok := s.groups[name]; ok {
		panic(fmt.Sprintf("name already used: %+q", name))
	}

	g := &group{clct: s, vals: counters}
	s.groups[name] = g
	g.AfterChange()
	return g
}

// Like AddNamed, but just for one counter
func (s *Service) Add(name string, counter *atomic.Uint64) *group {
	return s.AddNamed(name, map[string]*atomic.Uint64{"": counter})
}

// Returns true if there are any changes
func (s *Service) loadAllStats(values map[string]uint64) bool {
	s.mut.RLock()
	defer s.mut.RUnlock()

	changed := false

	for gName, g := range s.groups {
		func() {
			if g.mut.TryLock() {
				defer g.mut.Unlock()
				for valName, val := range g.vals {
					n := gName
					if valName != "" {
						n += "_" + valName
					}

					v := val.Load()
					if values[n] != v {
						changed = true
						values[n] = v
					}

				}
				if g.removed {
					delete(s.groups, gName)
				}
			}
		}()
	}

	return changed
}
