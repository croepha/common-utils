package grpctest_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/croepha/common-utils/grpctest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

type testServer struct {
	helloworld.UnimplementedGreeterServer
	t *testing.T
}

func (s *testServer) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	require.Equal(s.t, "myName", req.Name)
	return &helloworld.HelloReply{Message: "testMessage"}, nil
}

func TestBasic(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	server := grpc.NewServer()
	helloworld.RegisterGreeterServer(server, &testServer{t: t})

	client, err := grpc.NewClient("127.0.0.1",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpctest.StartTestGRPCTestServer(ctx, t, server),
	)
	require.NoError(t, err)

	resp, err := helloworld.NewGreeterClient(client).SayHello(
		context.Background(), &helloworld.HelloRequest{Name: "myName"})
	require.NoError(t, err)
	require.Equal(t, "testMessage", resp.Message)

}
