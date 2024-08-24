package grpctest

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

// Facilities to aid with GRPC unit tests

// Mock transport for GRPC:
// This sets up a grpc server using bufcon transport for testing., it takes in a server and it will start
// listening with that server, it will clean up with t.Cleanup or on the given ctx cancel
// This returns a DialOption that will set the Dialer to connect to this server, for use with NewClient
func StartTestGRPCTestServer(ctx context.Context, t *testing.T, baseServer *grpc.Server) grpc.DialOption {

	lis := bufconn.Listen(1 << 26)

	var testBlockers sync.WaitGroup
	// Prevent test from exiting until Serve goroutine has
	// completed.
	t.Cleanup(testBlockers.Wait)

	// trigger shutdown on either context or test done
	context.AfterFunc(ctx, baseServer.Stop)
	t.Cleanup(baseServer.Stop)

	// Serve in the background
	testBlockers.Add(1)
	go func() {
		defer testBlockers.Done()
		defer func() { require.NoError(t, lis.Close()) }()
		require.NoError(t, baseServer.Serve(lis))
	}()

	return grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
		// TODO: I'm thinking that injecting a timeout here might be a good idea
		// if we expanded use of this in a more general way
		return lis.DialContext(ctx)
	})
}
