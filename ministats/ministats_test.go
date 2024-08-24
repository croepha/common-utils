package ministats_test

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/ministats"
	"golang.org/x/exp/maps"
)

func TestBasic(t *testing.T) {

	deadlockWait := 100 * time.Millisecond

	extraEdgeCases := true

	statOut := make(chan map[string]uint64)

	expectedOut := map[string]uint64{}

	c := ministats.Service{
		PrintCooldown: time.Microsecond,
		PrintOutput: func(m map[string]uint64) {
			t.Log("statsOut:")
			keys := maps.Keys(m)
			slices.Sort(keys)
			for _, k := range keys {
				t.Logf("\t%s:\t%d", k, m[k])
			}
			statOut <- m
		},
	}

	requireStatOut := func(readsExpected int) {
		readsDoneBeforeTimeout := 0
		t.Helper()
		for {
			tmStart := time.Now()
			select {
			case <-time.After(deadlockWait):
				require.Equal(t, readsExpected, readsDoneBeforeTimeout)
				return
			case so := <-statOut:
				lockWaitTime := time.Since(tmStart)
				t.Logf("lockWaitTime: %s", lockWaitTime)
				require.Less(t, readsDoneBeforeTimeout, readsExpected)
				readsDoneBeforeTimeout++
				require.Equal(t, expectedOut, so)
				if lockWaitTime*100 > deadlockWait {
					t.Logf("WARNING: Lock wait time (%s) was within 1%% of deadlockWait time (%s)",
						lockWaitTime, deadlockWait,
					)
				}

			}
		}
	}

	//  Single value group
	g0v := atomic.Uint64{}
	g0v.Store(10)
	g0 := c.Add("group0", &g0v)
	expectedOut["group0"] = 10
	g0.AfterChange()

	//  Multi value group
	g1v0 := atomic.Uint64{}
	g1v0.Store(20)
	g1v1 := atomic.Uint64{}
	g1v1.Store(30)
	g1 := c.AddNamed("group1", map[string]*atomic.Uint64{
		"v0": &g1v0,
		"v1": &g1v1,
	})
	expectedOut["group1_v0"] = 20
	expectedOut["group1_v1"] = 30
	g1.AfterChange()

	// Not started, expect no output
	requireStatOut(0)

	c.Start()
	if extraEdgeCases {
		c.Start()
	}

	// Started, expect one from startup
	requireStatOut(1)

	// Add a new group
	g2v := atomic.Uint64{}
	g2v.Store(40)
	g2 := c.Add("group2", &g2v)
	expectedOut["group2"] = 40
	g2.AfterChange()
	requireStatOut(1)

	// Again but multivalue
	g3v0 := atomic.Uint64{}
	g3v0.Store(50)
	g3v1 := atomic.Uint64{}
	g3v1.Store(60)
	g3 := c.AddNamed("group3", map[string]*atomic.Uint64{
		"v0": &g3v0,
		"v1": &g3v1,
	})
	expectedOut["group3_v0"] = 50
	expectedOut["group3_v1"] = 60
	g3.AfterChange()
	requireStatOut(1)

	// Increment some random counters
	g0v.Add(100)
	g1v1.Add(100)
	g3v0.Add(100)
	expectedOut["group0"] = 110
	expectedOut["group1_v1"] = 130
	expectedOut["group3_v0"] = 150
	g0.AfterChange()
	g1.AfterChange()
	g3.AfterChange()
	requireStatOut(1)

	// Test BeforeChange preventing partial updates
	g1.BeforeChange()
	g1v0.Add(100)

	// Another update in the middle
	g0.BeforeChange()
	g0v.Add(100)
	expectedOut["group0"] = 210
	g0.AfterChange()
	requireStatOut(1)

	// Then finish out the g1 update
	g1v1.Add(100)
	expectedOut["group1_v0"] = 120
	expectedOut["group1_v1"] = 230
	g1.AfterChange()
	requireStatOut(1)

	// Do lots of updates within a long cooldown, ensure only one output
	c.PrintCooldown = time.Second // This is normally racey, but since
	// we have tight control, we can ensure this won't race
	g0v.Add(100)
	expectedOut["group0"] = 310
	g0.AfterChange()

	requireStatOut(1)
	midwayDeadLine := time.Now().Add(c.PrintCooldown / 2)
	afterDeadLine := time.Now().Add(c.PrintCooldown + 100*time.Millisecond)
	for range 5 {
		g0v.Add(100)
		g0.AfterChange()
	}
	expectedOut["group0"] = 810
	requireStatOut(0)
	<-time.After(time.Until(midwayDeadLine))
	requireStatOut(0)
	<-time.After(time.Until(afterDeadLine))
	requireStatOut(1)

	c.Stop()
	if extraEdgeCases {
		c.Stop()
	}

	requireStatOut(0)

}

// TODO: Add torture/scale test
