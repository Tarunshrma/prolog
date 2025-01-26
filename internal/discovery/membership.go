package discovery

import (
	"net"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
	"go.uber.org/zap"
)

type Membership struct {
	Config
	handler Handler
	serf    *serf.Serf
	events  chan serf.Event
	logger  *zap.Logger
}

func New(handler Handler, config Config) (*Membership, error) {
	m := &Membership{
		Config:  config,
		handler: handler,
		events:  make(chan serf.Event),
		logger:  zap.L().Named("membership"),
	}

	if err := m.setupSerf(); err != nil {
		return nil, err
	}

	return m, nil
}

type Config struct {
	NodeName       string
	BindAddr       string
	Tags           map[string]string
	StartJoinAddrs []string
}

func (m *Membership) setupSerf() error {
	addr, err := net.ResolveTCPAddr("tcp", m.Config.BindAddr)
	if err != nil {
		return err
	}
	config := serf.DefaultConfig()
	config.Init()
	config.MemberlistConfig.BindAddr = addr.IP.String()
	config.MemberlistConfig.BindPort = addr.Port

	//Is this step really needed, I am already setting the event channel in constructor
	m.events = make(chan serf.Event)

	config.EventCh = m.events
	config.NodeName = m.NodeName
	config.Tags = m.Tags

	serf, err := serf.Create(config)
	if err != nil {
		return err
	}

	m.serf = serf

	go m.eventHandler()
	if m.StartJoinAddrs != nil {
		_, err := m.serf.Join(m.StartJoinAddrs, true)
		if err != nil {
			return err
		}
	}

	return nil
}

type Handler interface {
	Join(name, addr string) error
	Leave(name string) error
}

func (m *Membership) eventHandler() {
	for e := range m.events {
		switch e.EventType() {
		case serf.EventMemberJoin:
			for _, member := range e.(serf.MemberEvent).Members { // e.(serf.MemberEvent) ??
				if m.isLocal(member) {
					continue
				}
				m.handleJoin(member)
			}
		case serf.EventMemberLeave, serf.EventMemberFailed:
			for _, member := range e.(serf.MemberEvent).Members {
				if m.isLocal(member) {
					continue
				}
				m.handleLeave(member)
			}
		}
	}
}

func (m *Membership) handleJoin(member serf.Member) {
	m.logger.Info("Node joined", zap.String("name", member.Name), zap.String("addr", member.Addr.String()))
	if err := m.handler.Join(member.Name, member.Tags["rpc_addrs"]); err != nil {
		m.logError(err, "Failed to handle join", member)
	}
}

func (m *Membership) handleLeave(member serf.Member) {
	m.logger.Info("Node left", zap.String("name", member.Name), zap.String("addr", member.Addr.String()))
	if err := m.handler.Leave(member.Name); err != nil {
		m.logError(err, "Failed to handle leave", member)
	}
}

func (m *Membership) isLocal(member serf.Member) bool {
	return member.Name == m.serf.LocalMember().Name
}

func (m *Membership) Members() []serf.Member {
	return m.serf.Members()
}

func (m *Membership) Leave() error {
	return m.serf.Leave()
}

func (m *Membership) logError(err error, msg string, member serf.Member) {
	log := m.logger.Error
	if err == raft.ErrNotLeader {
		log = m.logger.Debug
	}
	log(msg,
		zap.Error(err),
		zap.String("name", member.Name),
		zap.String("rpc_addr", member.Tags["rpc_addr"]))
}
