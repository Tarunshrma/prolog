package log

import (
	"io/ioutil"
	"os"
	"testing"

	api "github.com/Tarunshrma/prolog/log/api/v1"
	"github.com/test-go/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestLog(t *testing.T) {
	for scenario, fn := range map[string]func(t *testing.T, log *Log){
		"append and read a record succeeds": testStoreAppendRead,
		"offset out of range error":         testOutOfRangeErr,
		"init with existing segments":       testInitExisting,
		"reader":                            testReader,
		"truncate":                          testTruncate,
	} {
		t.Run(scenario, func(t *testing.T) {
			dir, err := ioutil.TempDir("", "store-test")
			require.NoError(t, err)
			defer os.RemoveAll(dir)

			c := Config{}
			c.Segment.MaxStoreBytes = 32
			log, err := NewLog(dir, c)
			require.NoError(t, err)

			fn(t, log)
		})
	}
}

func testStoreAppendRead(t *testing.T, log *Log) {
	record := &api.Record{
		Value: []byte("hello world"),
	}

	off, err := log.Append(record)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	got, err := log.Read(off)
	require.NoError(t, err)
	require.Equal(t, record.Value, got.Value)
}

func testOutOfRangeErr(t *testing.T, log *Log) {
	_, err := log.Read(1)
	require.Error(t, err)
}

func testInitExisting(t *testing.T, log *Log) {
	record := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		off, err := log.Append(record)
		require.NoError(t, err)
		require.Equal(t, uint64(i), off)
	}

	require.NoError(t, log.Close())

	off, err := log.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = log.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

	// Create a new log instance to simulate a restart
	log, err = NewLog(log.Dir, log.Config)
	require.NoError(t, err)

	off, err = log.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	off, err = log.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)

}

func testReader(t *testing.T, log *Log) {
	record := &api.Record{
		Value: []byte("hello world"),
	}

	off, err := log.Append(record)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	reader := log.Reader()
	b, err := ioutil.ReadAll(reader)
	require.NoError(t, err)

	read := &api.Record{}
	err = proto.Unmarshal(b[lenWidth:], read)
	require.NoError(t, err)
	require.Equal(t, record.Value, read.Value)
}

func testTruncate(t *testing.T, log *Log) {
	record := &api.Record{
		Value: []byte("hello world"),
	}

	for i := 0; i < 3; i++ {
		_, err := log.Append(record)
		require.NoError(t, err)
	}

	err := log.Truncate(1)
	require.NoError(t, err)

	off, err := log.LowestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(1), off)

	off, err = log.HighestOffset()
	require.NoError(t, err)
	require.Equal(t, uint64(2), off)
}
