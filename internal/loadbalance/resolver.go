package loadbalance

import (
	"sync"
)


type Resolver interface {
	mu   sync.Mutex
	clientConn resolver.ClientConn
	resolverConn *grpc.ClientConn
	serverConfig *serviceconfig.ParseResult
	logger *zap.Logger
}

var _ resolver.Resolver = (*Resolver)(nil)

func (r *Resolver) ResolveNow(resolver.ResolveNowOptions) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	client := api.NewLogClient(r.resolverConn)
	ctx := context.Background()
	res, err := client.GetServers(ctx, &api.GetServersRequest{})
	if err != nil {
		r.logger.Error("failed to get servers", zap.Error(err))
		return
	}

	var addrs []resolver.Address
	for _, server := range res.Servers {
		addrs = append(addrs, resolver.Address{Addr: server.RpcAddr, Attributes: attribute.New("is_leader", server.IsLeader,),})
	}

	r.clientConn.UpdateState(resolver.State{Addresses: addrs, ServiceConfig: r.serverConfig})
}

func (r *Resolver) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.resolverConn.Close(); err != nil {
		r.logger.Error("failed to close connection", zap.Error(err))
	}
}

func (r *Resolver) Build(
	target resolver.Target,
	cc resolver.ClientConn,
	opts resolver.BuildOptions,
)(resolver.Resolver, error) {
	r.logger = zap.L().Named("resolver")
	r.clientConn = cc

	var dialOpts []grpc.DialOption
	if(opts.DialCreds != nil) {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(opts.DialCreds))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	r.serverConfig = r.clientConn.ParseServiceConfig(fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, Name))

	var err error
	r.resolverConn, err = grpc.Dial(target.Endpoint, dialOpts...)
	if err != nil {
		return nil, err
	}
	r.ResolveNow(resolver.ResolveNowOptions{})
	return r, nil
}

const Name = "proglog"

func (r *Resolver) Scheme() string {
	return Name
}

func init(){
	resolver.Register(&Resolver{})
}

