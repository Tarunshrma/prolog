# Logging System Store Module

Welcome to the **Store Module** documentation for our logging system. This guide provides a comprehensive understanding of the `store` struct and its associated methods: `newStore`, `Append`, `Read`, `ReadAt`, and `Close`. Visual aids are included to help explain the concepts in a clear and intuitive way.

---

## Table of Contents

1. [Overview](#overview)
2. [Package and Imports](#package-and-imports)
3. [Global Variables](#global-variables)
4. [`store` Struct](#store-struct)
5. [Constructor: `newStore`](#constructor-newstore)
6. [Method: `Append`](#method-append)
7. [Method: `Read`](#method-read)
8. [Method: `ReadAt`](#method-readat)
9. [Method: `Close`](#method-close)
10. [Summary](#summary)
11. [Additional Considerations](#additional-considerations)
12. [Conclusion](#conclusion)

---

## Overview

The `store` module is an essential part of our logging system, managing the storage of log entries in an efficient and thread-safe manner. It allows for appending data, reading data at specific positions, and ensuring that the file is properly closed, maintaining data integrity.

**Key Features:**

- **Buffered Writing:** Uses `bufio.Writer` to reduce system calls and improve write performance.
- **Thread-Safe Access:** Uses mutex locks to ensure thread safety during read and write operations.
- **Length Prefix:** Stores the length of each data entry to facilitate efficient reading.

---

## Package and Imports

```go
package log

import (
    "bufio"
    "encoding/binary"
    "os"
    "sync"
)
```

- **Package `log`:** This module is part of the `log` package for managing log storage.
- **Imports:**
  - **`bufio`:** Provides buffered I/O to improve write performance.
  - **`encoding/binary`:** Used for encoding/decoding data in a binary format (BigEndian in this case).
  - **`os`:** Provides platform-independent file operations.
  - **`sync`:** Provides synchronization primitives like `Mutex` to ensure thread-safe access.

---

## Global Variables

```go
var (
    enc = binary.BigEndian
)

const (
    lenWidth = 8
)
```

- **`enc = binary.BigEndian`:** Specifies the encoding format for writing and reading the length of data entries.
- **`lenWidth = 8`:** Defines the width (in bytes) of the length prefix for each data entry.

**Visualization:**

```
Length Prefix (8 bytes):
+-------------------+
| Length of Data    |
| (8 bytes, BigEndian) |
+-------------------+
```

---

## `store` Struct

```go
type store struct {
    file *os.File
    mu   sync.Mutex
    buf  *bufio.Writer
    size uint64
}
```

- **Fields:**
  - **`file *os.File`:** Pointer to the file where data is stored.
  - **`mu sync.Mutex`:** Mutex to ensure thread-safe access.
  - **`buf *bufio.Writer`:** Buffered writer to wrap the `os.File` object and improve write performance.
  - **`size uint64`:** Tracks the current size of the store file in bytes.

**Visualization:**

```
store Struct:
+------+---------+-----------+------+
| File | Mutex   | Buffer    | Size |
+------+---------+-----------+------+
| *os.File | sync.Mutex | *bufio.Writer | uint64 |
+------+---------+-----------+------+
```

---

## Constructor: `newStore`

```go
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
```

### Purpose

Creates a new `store` instance that wraps an existing file. It sets up the buffered writer and calculates the current file size.

### Parameters

- **`f *os.File`:** The file to be used for storing log entries.

### Returns

- **`*store`:** Pointer to the initialized `store` instance.
- **`error`:** Error object if initialization fails.

### Step-by-Step Breakdown

1. **Get File Information:**

   ```go
   fi, err := os.Stat(f.Name())
   if err != nil {
       return nil, err
   }
   ```
   - **Purpose:** Retrieve information about the file to determine its size.

2. **Initialize `store` Instance:**

   ```go
   size := uint64(fi.Size())

   return &store{
       file: f,
       size: size,
       buf:  bufio.NewWriter(f),
   }, nil
   ```
   - **Purpose:** Create and return the `store` instance with the appropriate size and buffer.

**Visualization:**

```
Initialization Flow:
+---------------------+
|    newStore(f)      |
+---------------------+
           |
           v
+---------------------+
| Get file info       |
| Set size            |
+---------------------+
           |
           v
+---------------------+
| Initialize store    |
+---------------------+
           |
           v
+---------------------+
| Return store instance|
+---------------------+
```

---

## Method: `Append`

```go
func (s *store) Append(p []byte) (n uint64, pos uint64, err error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    pos = s.size

    if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
        return 0, 0, err
    }

    w, err := s.buf.Write(p)
    if err != nil {
        return 0, 0, err
    }

    w += lenWidth
    s.size += uint64(w)

    return uint64(w), pos, nil
}
```

### Purpose

Appends the given byte slice to the store. The length of the data is written first, followed by the actual data.

### Parameters

- **`p []byte`**: The byte slice to be appended to the store.

### Returns

- **`n uint64`**: The number of bytes written, including the length prefix.
- **`pos uint64`**: The position where the data was written.
- **`err error`**: Error object if the write operation fails.

### Step-by-Step Breakdown

1. **Acquire Mutex Lock:**

   ```go
   s.mu.Lock()
   defer s.mu.Unlock()
   ```
   - **Purpose:** Ensure thread-safe access by locking the store during the write operation.

2. **Set Position:**

   ```go
   pos = s.size
   ```
   - **Purpose:** Record the position where the new data will be written.

3. **Write Length Prefix:**

   ```go
   if err := binary.Write(s.buf, enc, uint64(len(p))); err != nil {
       return 0, 0, err
   }
   ```
   - **Purpose:** Write the length of the data to the buffer.

4. **Write Data:**

   ```go
   w, err := s.buf.Write(p)
   if err != nil {
       return 0, 0, err
   }
   ```
   - **Purpose:** Write the actual data to the buffer.

5. **Update Size:**

   ```go
   w += lenWidth
   s.size += uint64(w)
   ```
   - **Purpose:** Update the store size to reflect the new data.

6. **Return Result:**

   ```go
   return uint64(w), pos, nil
   ```
   - **Purpose:** Return the number of bytes written and the position.

**Visualization:**

```
Append Operation:
+----------------------+
| Call Append(p []byte)|
+----------------------+
           |
           v
+----------------------+
| Acquire lock         |
+----------------------+
           |
           v
+----------------------+
| Set position (pos)   |
+----------------------+
           |
           v
+----------------------+
| Write length prefix  |
+----------------------+
           |
           v
+----------------------+
| Write data           |
+----------------------+
           |
           v
+----------------------+
| Update size          |
+----------------------+
           |
           v
+----------------------+
| Return result        |
+----------------------+
```

---

## Method: `Read`

```go
func (s *store) Read(pos uint64) ([]byte, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if err := s.buf.Flush(); err != nil {
        return nil, err
    }

    size := make([]byte, lenWidth)
    if _, err := s.file.ReadAt(size, int64(pos)); err != nil {
        return nil, err
    }

    b := make([]byte, enc.Uint64(size))
    if _, err := s.file.ReadAt(b, int64(pos+lenWidth)); err != nil {
        return nil, err
    }

    return b, nil
}
```

### Purpose

Reads data from the store at the specified position. It first reads the length of the data and then reads the actual data.

### Parameters

- **`pos uint64`**: The position to start reading from.

### Returns

- **`[]byte`**: The data read from the store.
- **`error`**: Error object if the read operation fails.

### Step-by-Step Breakdown

1. **Acquire Mutex Lock:**

   ```go
   s.mu.Lock()
   defer s.mu.Unlock()
   ```
   - **Purpose:** Ensure thread-safe access by locking the store during the read operation.

2. **Flush Buffer:**

   ```go
   if err := s.buf.Flush(); err != nil {
       return nil, err
   }
   ```
   - **Purpose:** Flush any buffered writes to ensure data consistency.

3. **Read Length Prefix:**

   ```go
   size := make([]byte, lenWidth)
   if _, err := s.file.ReadAt(size, int64(pos)); err != nil {
       return nil, err
   }
   ```
   - **Purpose:** Read the length of the data at the given position.

4. **Read Data:**

   ```go
   b := make([]byte, enc.Uint64(size))
   if _, err := s.file.ReadAt(b, int64(pos+lenWidth)); err != nil {
       return nil, err
   }
   ```
   - **Purpose:** Read the actual data from the store.

5. **Return Data:**

   ```go
   return b, nil
   ```
   - **Purpose:** Return the data that was read.

**Visualization:**

```
Read Operation:
+----------------------+
| Call Read(pos uint64)|
+----------------------+
           |
           v
+----------------------+
| Acquire lock         |
+----------------------+
           |
           v
+----------------------+
| Flush buffer         |
+----------------------+
           |
           v
+----------------------+
| Read length prefix   |
+----------------------+
           |
           v
+----------------------+
| Read data            |
+----------------------+
           |
           v
+----------------------+
| Return data          |
+----------------------+
```

---

## Method: `ReadAt`

```go
func (s *store) ReadAt(p []byte, off int64) (int, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    return s.file.ReadAt(p, off)
}
```

### Purpose

Reads data from the store at a specified offset without considering the length prefix.

### Parameters

- **`p []byte`**: The buffer to read data into.
- **`off int64`**: The offset to start reading from.

### Returns

- **`int`**: The number of bytes read.
- **`error`**: Error object if the read operation fails.

### Step-by-Step Breakdown

1. **Acquire Mutex Lock:**

   ```go
   s.mu.Lock()
   defer s.mu.Unlock()
   ```
   - **Purpose:** Ensure thread-safe access by locking the store during the read operation.

2. **Read Data:**

   ```go
   return s.file.ReadAt(p, off)
   ```
   - **Purpose:** Read data from the store at the specified offset.

---

## Method: `Close`

```go
func (s *store) Close() error {
    s.mu.Lock()
    defer s.mu.Unlock()

    if err := s.buf.Flush(); err != nil {
        return err
    }

    return s.file.Close()
}
```

### Purpose

Closes the store, ensuring all buffered writes are flushed to the file before closing it.

### Returns

- **`error`**: Error object if the close operation fails.

### Step-by-Step Breakdown

1. **Acquire Mutex Lock:**

   ```go
   s.mu.Lock()
   defer s.mu.Unlock()
   ```
   - **Purpose:** Ensure thread-safe access by locking the store during the close operation.

2. **Flush Buffer:**

   ```go
   if err := s.buf.Flush(); err != nil {
       return err
   }
   ```
   - **Purpose:** Flush any buffered writes to ensure all data is written to disk.

3. **Close File:**

   ```go
   return s.file.Close()
   ```
   - **Purpose:** Close the underlying file.

**Visualization:**

```
Close Operation:
+----------------------+
| Call Close()         |
+----------------------+
           |
           v
+----------------------+
| Acquire lock         |
+----------------------+
           |
           v
+----------------------+
| Flush buffer         |
+----------------------+
           |
           v
+----------------------+
| Close file           |
+----------------------+
```

---

## Summary

The `store` module provides the core functionality for managing log data in our logging system. It ensures efficient storage and retrieval of data while maintaining thread safety and data consistency.

### Key Concepts

- **Buffered Writing:** Improves performance by reducing the number of system calls.
- **Length Prefix:** Allows for easy parsing and retrieval of data.
- **Thread Safety:** Mutexes are used to ensure that concurrent access to the store does not lead to data corruption.

---

## Additional Considerations

### 1. Concurrency Control

- **Mutex Usage:** Mutex locks (`s.mu`) are used to ensure thread safety for all read, write, and close operations.

### 2. Error Handling

- **Error Propagation:** All methods return error objects to indicate failures, allowing the caller to handle them appropriately.

### 3. Performance Optimization

- **Buffered Writes:** Using `bufio.Writer` helps minimize system calls, improving write performance.

---

## Conclusion

The `store` module is a fundamental component of our logging system that ensures efficient data management with thread safety and performance optimizations. By maintaining proper buffering, length-prefix encoding, and mutex-controlled access, it guarantees data integrity and reliability.

For any further questions or clarifications, feel free to refer to the extended documentation or reach out to the development team.

---

*End of Store Module Documentation*
