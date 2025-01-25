package log

import (
	"io"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
)

type DistributedLog struct {
	config Config
	log    *Log
	raft   *raft.Raft
}


func NewDistributedLog(dataDir string. config Config) (*DistributedLog, error) {
	l := &DistributedLog{
		config: config,
	}

	if err := l.setupLog(dataDir); err != nil {
		return nil, err
	}

	if err := l.setupRaft(dataDir); err != nil {
		return nil, err
	}

	return l, nil
}

func (l *DistributedLog) setupLog(dataDir string) error {
	logDir := filepath.Join(dataDir, "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	var err error
	l.log, err = NewLog(logDir, l.config)
	return err
}

func (l *DistributedLog) setupRaft(dataDir string) error {
	fsm := &fsm{log: l.log}

	logDir := filepath.Join(dataDir, "raft", "log")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	logConfig := l.config
	logConfig.Segment.InitialOffset =1
	logStore, err := newLogStore(logDir, logConfig)
	if err != nil {
		return err
	}

	//Key-value store where where raft store its metadata like current term, voted for etc. 
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft", "stable"))
	if err != nil {
		return err
	}

	retain := 1

	//Snapshot store where raft store snapshots
	snapshotStore, err := raft.NewFileSnapshotStore(
						filepath.Join(dataDir, "raft"),
						retain, os.Stderr)
	if err != nil {
		return err
	}

	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(
					l.config.Raft.StreamLayer, 
					maxPool, 
					timeout, 
					os.Stderr)

	config := raft.DefaultConfig()
	config.LocalID = l.config.Raft.LocalID	//raft.ServerID(l.config.Raft.LocalID)
	if l.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = l.config.Raft.HeartbeatTimeout
	}
	if l.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = l.config.Raft.ElectionTimeout
	}
	if l.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = l.config.Raft.LeaderLeaseTimeout
	}
	if l.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = l.config.Raft.CommitTimeout
	}						

	l.raft, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return err
	}
	hasState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		return err
	}
	if l.config.Raft.Bootstrap && !hasState {
		config := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      config.LocalID,
					Address: transport.LocalAddr(),
				},
			},
		}
		err = l.raft.BootstrapCluster(config).Error()
	} 

	return err
}

func (l *DistributedLog) Append(record *Record) (uint64, error) {
	res, err := l.apply(
		AppendRequestType,
		&api.ProduceRequest{Record: record},
	)

	if err != nil {
		return 0, nil
	}
	return res.(*AppendResponse).Offset, nil
}

func (l *DistributedLog) apply(reqType RequestType, req proto.Message) (interface{}, error) {
	var buf bytes.Buffer
	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		return nil, err
	}
	
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}
	timeout := 10 * time.Second
	f := l.raft.Apply(buf.Bytes(), timeout)
	if f.Error() != nil {
		return nil, f.Error()
	}
	res := f.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res, nil
}

func (l *DistributedLog) Read(offset uint64) (*Record, error) {
	return l.log.Read(offset)
}
var _ raft.FSM = (*fsm)(nil)

type fsm struct {
	log *Log
}

type RequestType uint8

const (
	AppendRequestType RequestType = 0
)

func (l *fsm) Apply(record *raft.Log) interface{} {
	buf := record.Data
	reqType := RequestType(buf[0])
	switch reqType {
	case AppendRequestType:
		return l.applyAppend(buf[1:])
	}
	return nil
}

func (l *fsm) applyAppend(b []byte) interface{} {
	var req api.ProduceRequest
	err := proto.Unmarshal(b, &req)

	if err != nil {
		return err
	}

	offset, err := l.log.Append(req.Record)
	if err != nil {
		return err
	}

	return &api.ProduceResponse{Offset: offset}
}

func (l *fsm) Snapshot() (raft.FSMSnapshot, error) {
	r := l.log.Reader()
	return &snapshot{reader: r}, nil
}

var _ raft.FSMSnapshot = (*snapshot)(nil)

type snapshot struct {
	reader io.Reader
}

func (s *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := io.Copy(sink, s.reader); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (s *snapshot) Release() {}

func (l *fsm) Restore(r io.ReadCloser) error {	
	b := make([]byte, lenWidth)
	var buf bytes.Buffer

	for i := 0; ; i++ {
		_, err := io.ReadFull(r,b)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		size := int64(enc.Uint64(b))
		if _, err = io.CopyN(&buf, r, size); err != nil {
			return err
		}

		record := &api.Record{}
		if err = proto.Unmarshal(buf.Bytes(), record); err != nil {
			return err
		}

		if i ==0 {
			f.log.Config.Segment.InitialOffset = record.Offset
			if err = f.log.Reset(); err != nil {
				return err
			}
		}

		if _, err = f.log.Append(record); err != nil {
			return err
		}

		buf.Reset()
	}

	return nil
}

var _ raft.LogStore = (*logStore)(nil)

type logStore struct {
	*Log
}

fun newLogStore(dir string, config Config) (*logStore, error) {
	l, err := NewLog(dir, config)
	if err != nil {
		return nil, err
	}
	return &logStore{Log: l}, nil
}

func (s *logStore) FirstIndex() (uint64, error) {
	return s.FirstOffset()
}

func (s *logStore) LastIndex() (uint64, error) {
	off, err := s.HighestOffset()
	return off, err
}

func (s *logStore) GetLog(index uint64, out *raft.Log) error {
	in, err := s.Read(index)
	if err != nil {
		return err
	}

	out.Data = in.Value
	out.Index = in.Offset
	out.Term = in.Term
	return nil
}

func (s *logStore) StoreLog(log *raft.Log) error {
	return s.StoreLogs([]*raft.Log{log})
}

func (s *logStore) StoreLogs(logs []*raft.Log) error {
	for _, l := range logs {
		if _, err := s.Append(&Record{
			Term: l.Term,
			Value: l.Data,
			Type: uint32(l.Type),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *logStore) DeleteRange(min, max uint64) error {
	return s.Truncate(min)
}