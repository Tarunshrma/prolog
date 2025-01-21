package log

import (
	"context"
	"sync"

	api "github.com/Tarunshrma/prolog/api/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Replicator struct {
	// Replicate the given log entry to all peers.
	DialOptions []grpc.DialOption
	LocalServer api.LogClient

	//using refrence type nsures that all parts of your program referencing the logger are accessing the same instance and its state.
	logger *zap.Logger

	mu sync.Mutex

	//servers is a map of all the servers that are currently connected to the replicator.
	//It is used to stop replicating to a server when it is closed or failed.
	servers map[string]chan struct{}

	closed bool
	close  chan struct{}
}

func (r *Replicator) Join(name, addrs string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.init()

	if r.closed {
		return nil
	}

	if _, ok := r.servers[name]; ok {
		//server already exists
		return nil
	}

	r.servers[name] = make(chan struct{})
	go r.replicate(addrs, r.servers[name])
	return nil
}

func (r *Replicator) replicate(addrs string, leave chan struct{}) {
	cc, err := grpc.Dial(addrs, r.DialOptions...)
	if err != nil {
		r.logger.Error("failed to dial", zap.String("address", addrs), zap.Error(err))
		return
	}
	defer cc.Close()

	//grpc client
	client := api.NewLogClient(cc)
	ctx := context.Background()
	stream, err := client.ConsumeStream(ctx,
		&api.ConsumeRequest{
			Offset: 0,
		})

	if err != nil {
		r.logger.Error("failed to consume", zap.Error(err))
		return
	}

	records := make(chan *api.Record)
	go func() {
		for {
			resp, err := stream.Recv()
			if err != nil {
				r.logger.Error("failed to receive", zap.Error(err))
				return
			}
			records <- resp.Record
		}
	}()

	for {
		select {
		case <-r.close:
			return
		case <-leave:
			return
		case record := <-records:
			_, err := r.LocalServer.Produce(ctx,
				&api.ProduceRequest{
					Record: record,
				})
			if err != nil {
				r.logger.Error("failed to produce", zap.Error(err))
				return
			}
		}
	}
}

func (r *Replicator) Leave(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.init()

	if _, ok := r.servers[name]; !ok {
		return nil
	}

	close(r.servers[name])
	delete(r.servers, name)

	return nil
}

func (r *Replicator) init() {
	if r.logger == nil {
		r.logger = zap.L().Named("replicator")
	}

	if r.servers == nil {
		r.servers = make(map[string]chan struct{})
	}

	if r.close == nil {
		r.close = make(chan struct{})
	}
}

func (r *Replicator) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.init()

	if r.closed {
		return nil
	}

	r.closed = true
	close(r.close)

	return nil
}
