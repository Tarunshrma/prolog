package log

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc = binary.BigEndian
)

const (
	lenWidth = 8
)

type store struct {
	file *os.File
	mu   sync.Mutex

	//A buffered writer that wraps the os.File object.
	//Buffering improves write performance by minimizing the number of system calls.
	//instead of writing directly to the file, you're writing to a buffer, and then the buffer writes to the file.
	buf  *bufio.Writer
	size uint64
}

// newStore creates a new store object.
func newStore(f *os.File) (*store, error) {
	fi, err := os.Stat(f.Name())
	if err != nil {
		return nil, err
	}

	size := uint64(fi.Size())

	return &store{
		file: f,
		size: size,
		buf:  bufio.NewWriter(f),
	}, nil
}

// Append appends the provided byte slice to the store.
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
	// Acquire the lock to ensure thread-safe access to the store.
	s.mu.Lock()
	defer s.mu.Unlock() // Release the lock when the function exits.

	pos = s.size

	/* Why Write the Length First */
	/*
	* Writing the length of the data before the actual data allows for easier reading and parsing later.
	* When reading, you can first read the length, know exactly how many bytes to read for the data, and process accordingly.
	 */
	if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
		return 0, 0, err
	}

	// It returns the number of bytes written
	w, err := s.buf.Write(p)

	if err != nil {
		return 0, 0, err
	}

	w += lenWidth
	s.size += uint64(w)

	return uint64(w), pos, nil
}

func (s *store) Read(pos uint64) ([]byte, error) {
	// Acquire the lock to ensure thread-safe access to the store.
	s.mu.Lock()
	defer s.mu.Unlock() // Release the lock when the function exits.

	// Flush the buffer to ensure that any buffered writes are committed to the file.
	if err := s.buf.Flush(); err != nil {
		return nil, err // If flushing the buffer fails, return an error.
	}

	// Create a slice of bytes to hold the size information (length of the data).
	// lenWidth is assumed to be a predefined constant representing the size of the length header.

	//size = 00 00 00 00 00 00 00 00
	// where it is 8 bytes defined by lenWidth
	size := make([]byte, lenWidth)

	// Read the length of the stored data from the file.
	// The length is stored at the position indicated by `pos`.
	// ReadAt reads len(size) bytes starting from `pos`.

	//size = 00 00 00 00 00 00 00 05
	//where 05 is the length of the data e.f. Hello
	if _, err := s.file.ReadAt(size, int64(pos)); err != nil {
		return nil, err // Return an error if reading the length fails.
	}

	// Decode the size using BigEndian encoding.
	// `enc.Uint64(size)` reads the size from the byte slice `size` (which contains the length prefix).

	//b = 00 00 00 00 00
	//5 bytes of data as read previosly
	b := make([]byte, enc.Uint64(size))

	// Read the actual data from the file.
	// The position for reading starts after the length prefix (`pos + lenWidth`).

	//b = H e l l o
	// binary representation of ascii data where each letter reprent a byte.
	if _, err := s.file.ReadAt(b, int64(pos+lenWidth)); err != nil {
		return nil, err // Return an error if reading the actual data fails.
	}

	// Return the read data and a nil error indicating success.
	return b, nil
}

func (s *store) ReadAt(p []byte, off int64) (int, error) {
	// Acquire the lock to ensure thread-safe access to the store.
	s.mu.Lock()
	defer s.mu.Unlock() // Release the lock when the function exits.

	// Read the actual data from the file.
	// The position for reading starts after the length prefix (`pos + lenWidth`).
	return s.file.ReadAt(p, off)
}

func (s *store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.buf.Flush(); err != nil {
		return err
	}

	return s.file.Close()
}

func (s *store) Name() string {
	return s.file.Name()
}
