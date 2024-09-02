package minicgroups_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/logging"
	"gitlab.com/croepha/common-utils/lostandfound"
	"gitlab.com/croepha/common-utils/minicgroups"
	"gitlab.com/croepha/common-utils/syscallextra"
)

func TestSuccess(t *testing.T) {
	ctx := context.Background()

	cgm, err :=  minicgroups.NewMount(ctx)
	require.NoError(t, err)

	defer func() {
		err := cgm.Done(ctx)
		require.NoError(t, err)
		require.Panics(t, func() {
			cgm.Done(ctx)
		})
	}()

	g, err := cgm.CreateGroup(ctx, []string{"memory", "cpu"})
	require.NoError(t, err)

	err = g.WriteFiles(ctx, map[string]string{
		"memory.max": "134217728",
		"cpu.max":    "1000 100000",
	})
	require.NoError(t, err)

	fcs, err := g.ReadFiles(ctx, []string{
		"memory.max",
		"cpu.max",
	})
	require.NoError(t, err)
	require.Equal(t, []string{"134217728\n", "1000 100000\n"}, fcs)

	for fd, err := range g.FD(ctx) {
		require.NoError(t, err)
		fdstr, err := syscallextra.DescribeOpenFD(fd)
		require.NoError(t, err)
		defer syscallextra.RequiredFDClosed(t, fd, fdstr)
		lostandfound.RequireFileContent(t, fdstr+"/memory.max", "134217728")
	}

}

func TestMain(m *testing.M) {
	logging.SlogStartup()
	m.Run()
}
