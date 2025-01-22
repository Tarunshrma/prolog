package agent_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	api "github.com/Tarunshrma/prolog/api/v1"
	"github.com/Tarunshrma/prolog/internal/agent"
	"github.com/test-go/testify/require"
	"github.com/travisjeffery/go-dynaport"
	"google.golang.org/grpc"
)

func TestAgent(t *testing.T) {
	var agents []*agent.Agent
	for i := 0; i < 3; i++ {
		ports := dynaport.Get(2)
		bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
		rpcPort := ports[1]

		dataDir, err := ioutil.TempDir("", "agent-test")
		require.NoError(t, err)

		var startJoinAddrs []string
		if i != 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr)
		}

		a, err := agent.New(agent.Config{
			NodeName:       fmt.Sprintf("%d", i),
			StartJoinAddrs: startJoinAddrs,
			BindAddr:       bindAddr,
			RPCPort:        rpcPort,
			DataDir:        dataDir,
		})
		require.NoError(t, err)

		agents = append(agents, a)
	}

	defer func() {
		for _, a := range agents {
			err := a.Shutdown()
			require.NoError(t, err)
			require.NoError(t, os.RemoveAll(a.Config.DataDir))
		}
	}()
	time.Sleep(3 * time.Second)

	leaderClient := client(t, agents[0])
	produceResp, err := leaderClient.Produce(context.Background(), &api.ProduceRequest{
		Record: &api.Record{
			Value: []byte("hello"),
		},
	},
	)

	require.NoError(t, err)
	consumeResp, err := leaderClient.Consume(context.Background(), &api.ConsumeRequest{
		Offset: produceResp.Offset,
	},
	)
	require.NoError(t, err)
	require.Equal(t, "hello", string(consumeResp.Record.Value))

	//wait untill the replication has finished
	time.Sleep(3 * time.Second)

	followerClient := client(t, agents[1])
	consumeResp, err = followerClient.Consume(context.Background(), &api.ConsumeRequest{
		Offset: produceResp.Offset,
	},
	)
	require.NoError(t, err)
	require.Equal(t, "hello", string(consumeResp.Record.Value))

}

func client(t *testing.T, a *agent.Agent) api.LogClient {
	rpcAddr, err := a.Config.RPCAddr()
	require.NoError(t, err)

	conn, err := grpc.Dial(fmt.Sprintf("%s", rpcAddr), grpc.WithInsecure())
	require.NoError(t, err)
	client := api.NewLogClient(conn)
	return client
}
