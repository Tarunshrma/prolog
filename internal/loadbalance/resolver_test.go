package loadbalance_test

import (
	"net"
	"testing"

	api "github.com/Tarunshrma/prolog/api/v1"
	loadbalance "github.com/Tarunshrma/prolog/internal/loadbalance"
	"github.com/Tarunshrma/prolog/internal/server"
	"github.com/test-go/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

func TestResolver(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1")
	require.NoError(t, err)

	srv, err := server.NewGRPCServer(&server.Config{
		GetServer: &getServer{},
	}, grpc.WithInsecure())

	require.NoError(t, err)
	go srv.Serve(l)

	conn := &clientConn{}
	r := &loadbalance.Resolver{}

	_, err = r.Build(
		resolver.Target{
			Endpoint: l.Addr().String(),
		},
		conn,
		resolver.BuildOptions{},
	)
	require.NoError(t, err)

	wantState := resolver.State{
		Addresses: []resolver.Address{{
			Addr:       "localhost:9001",
			Attributes: attributes.New("is_leader", true),
		},
			{
				Addr:       "localhost:9002",
				Attributes: attributes.New("is_leader", false),
			}},
	}
	require.Equal(t, wantState, conn.state)

	conn.state.Addresses = nil
	r.ResolveNow(resolver.ResolveNowOptions{})
	require.Equal(t, wantState, conn.state)
}

type getServer struct{}

func (g *getServer) GetServers() ([]*api.Server, error) {
	return []*server.Server{
		{
			Id:       "leader",
			RpcAddr:  "localhost:9001",
			IsLeader: true,
		},
		{
			Id:       "follower",
			RpcAddr:  "localhost:9002",
			IsLeader: false,
		},
	}, nil
}

type clientConn struct {
	resolver.ClientConn
	state resolver.State
}

func (c *clientConn) UpdateState(state resolver.State) {
	c.state = state
}

func (c *clientConn) ReportError(err error) {}

func (c *clientConn) NewAddress(addrs []resolver.Address) {}

func (c *clientConn) NewServiceConfig(serviceConfig string) {}

func (c *clientConn) ParseServiceConfig(config string) *serviceConfig.ParseResult {
	return nil
}
