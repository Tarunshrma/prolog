# Logging System Index Module

Welcome to the **Index Module** documentation for our high-performance logging system. This guide provides a comprehensive understanding of the `index` struct and its associated methods: `newIndex`, `Read`, `Write`, and `Close`. Visual aids are included to facilitate quick comprehension and serve as a handy reference for future reviews.

---

## Table of Contents

1. [Overview](#overview)
2. [Package and Imports](#package-and-imports)
3. [Global Variables](#global-variables)
4. [`index` Struct](#index-struct)
5. [Constructor: `newIndex`](#constructor-newindex)
6. [Method: `Read`](#method-read)
7. [Method: `Write`](#method-write)
8. [Method: `Close`](#method-close)
9. [Method: `Name`](#method-name)
10. [Visual Examples](#visual-examples)
11. [Summary](#summary)
12. [Additional Considerations](#additional-considerations)
13. [Conclusion](#conclusion)

---

## Overview

The `index` module is a critical component of our logging system, responsible for maintaining a mapping between **offsets** and **positions** of log entries. This mapping enables efficient retrieval of log data by allowing direct access to log entries in the store file without scanning the entire file.

**Key Features:**

- **Memory-Mapped Files:** Utilizes `gommap` for high-performance read/write operations.
- **Fixed-Size Index Entries:** Each index entry consists of an `offset` and a `position`.
- **Efficient Data Management:** Supports quick appends and reads, ensuring scalability for large log files.

---

## Package and Imports

```go
package log

import (
    "io"
    "os"

    "github.com/tysonmote/gommap"
)
```

- **Package `log`:** Indicates that this module is part of the `log` package, handling logging functionalities.
- **Imports:**
  - **`io`:** Provides basic interfaces for I/O operations.
  - **`os`:** Facilitates platform-independent file operations.
  - **`github.com/tysonmote/gommap`:** Enables memory-mapped file operations for efficient file access.

---

## Global Variables

```go
var (
    offWidth uint64 = 4
    posWidth uint64 = 8
    entWidth uint64 = offWidth + posWidth
)
```

- **Purpose:** Define the byte widths for the components of an index entry.
  
- **Variables:**
  - **`offWidth` (`4` bytes):** Width of the **offset** field (`uint32`).
  - **`posWidth` (`8` bytes):** Width of the **position** field (`uint64`).
  - **`entWidth` (`12` bytes):** Total width of an index entry (`offset + position`).

**Visualization:**

```
Index Entry (12 bytes):
+----------+----------+
| Offset   | Position |
| (4 bytes)| (8 bytes)|
+----------+----------+
```

---

## `index` Struct

```go
type index struct {
    file *os.File
    mmap gommap.MMap
    size uint64
}
```

- **Fields:**
  - **`file *os.File`:** Pointer to the index file on disk.
  - **`mmap gommap.MMap`:** Memory-mapped region of the index file for efficient access.
  - **`size uint64`:** Current size of the index data in bytes.

**Visualization:**

```
index Struct:
+------+-----------+------+
| File | MemoryMap | Size |
+------+-----------+------+
| *os.File | gommap.MMap | uint64 |
+------+-----------+------+
```

---

## Constructor: `newIndex`

```go
func newIndex(f *os.File, c Config) (*index, error) {
    idx := &index{
        file: f,
    }

    fi, err := os.Stat(f.Name())
    if err != nil {
        return nil, err
    }

    idx.size = uint64(fi.Size())
    if err := os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
        return nil, err
    }

    if idx.mmap, err = gommap.Map(
        idx.file.Fd(),
        gommap.PROT_READ|gommap.PROT_WRITE,
        gommap.MAP_SHARED,
    ); err != nil {
        return nil, err
    }

    return idx, nil
}
```

### Purpose

Initializes a new `index` instance by:

1. **Opening the Index File:** Associates the provided file handle.
2. **Setting Current Size:** Determines the existing size of the index file.
3. **Truncating the File:** Resizes the file to `MaxIndexBytes` to reserve space.
4. **Memory Mapping:** Maps the file into memory for efficient access.

### Parameters

- **`f *os.File`:** Open file handle for the index file.
- **`c Config`:** Configuration object containing settings like `MaxIndexBytes`.

### Returns

- **`*index`:** Pointer to the initialized `index` instance.
- **`error`:** Error object if initialization fails.

### Step-by-Step Breakdown

1. **Initialize `index` Instance:**

   ```go
   idx := &index{
       file: f,
   }
   ```

2. **Retrieve File Information:**

   ```go
   fi, err := os.Stat(f.Name())
   if err != nil {
       return nil, err
   }
   ```

3. **Set Current Size:**

   ```go
   idx.size = uint64(fi.Size())
   ```

4. **Truncate the File to Maximum Size:**

   ```go
   if err := os.Truncate(f.Name(), int64(c.Segment.MaxIndexBytes)); err != nil {
       return nil, err
   }
   ```

5. **Memory Map the File:**

   ```go
   if idx.mmap, err = gommap.Map(
       idx.file.Fd(),
       gommap.PROT_READ|gommap.PROT_WRITE,
       gommap.MAP_SHARED,
   ); err != nil {
       return nil, err
   }
   ```

6. **Return the Initialized `index`:**

   ```go
   return idx, nil
   ```

**Visualization:**

```
Initialization Flow:
+---------------------+
|    newIndex(f, c)   |
+---------------------+
           |
           v
+---------------------+
| Initialize index    |
| Set file to f       |
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
| Truncate file to    |
| MaxIndexBytes       |
+---------------------+
           |
           v
+---------------------+
| Memory map the file |
+---------------------+
           |
           v
+---------------------+
| Return index instance|
+---------------------+
```

---

## Method: `Read`

```go
func (i *index) Read(in int64) (out int32, pos uint64, err error) {
    if i.size == 0 {
        return 0, 0, nil
    }

    if in == -1 {
        out = int32((i.size / entWidth) - 1)
    } else {
        out = int32(in)
    }

    pos = uint64(out) * entWidth
    if i.size < pos+entWidth {
        return 0, 0, io.EOF
    }
    out = enc.Uint32(i.mmap[pos : pos+offWidth])
    pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
    return out, pos, nil
}
```

### Purpose

Retrieves an index entry based on the provided **offset**. It returns the corresponding **offset** and **position** from the index file.

### Parameters

- **`in int64`**: The input offset to read. Special case: `-1` indicates a request to read the **last** index entry.

### Returns

- **`out int32`**: The retrieved offset.
- **`pos uint64`**: The position in the store file corresponding to the offset.
- **`err error`**: Error object if the read operation fails.

### Step-by-Step Breakdown

1. **Check if Index is Empty:**

   ```go
   if i.size == 0 {
       return 0, 0, nil
   }
   ```

   - **Action:** If the index size is `0`, return zeros with no error.

2. **Determine Offset to Read:**

   ```go
   if in == -1 {
       out = int32((i.size / entWidth) - 1)
   } else {
       out = int32(in)
   }
   ```

   - **Scenario 1:** `in == -1`
     - **Action:** Set `out` to the last valid offset.
     - **Calculation:** `(i.size / entWidth) - 1`
   
   - **Scenario 2:** `in != -1`
     - **Action:** Set `out` to the provided offset.

3. **Calculate Byte Position in Index File:**

   ```go
   pos = uint64(out) * entWidth
   ```

   - **Action:** Determine where the index entry starts in the memory-mapped file.

4. **Validate Byte Position:**

   ```go
   if i.size < pos+entWidth {
       return 0, 0, io.EOF
   }
   ```

   - **Action:** Ensure the calculated position doesn't exceed the index size.
   - **Return:** `io.EOF` if out of bounds.

5. **Read Offset and Position from Index Entry:**

   ```go
   out = enc.Uint32(i.mmap[pos : pos+offWidth])
   pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth])
   return out, pos, nil
   ```

   - **Action:** Extract and decode the offset and position.
   - **Assumption:** `enc` is a predefined `binary.ByteOrder` (e.g., `binary.BigEndian`).

**Visualization:**

```
Read Operation:
+----------------------+
| Call Read(in int64)  |
+----------------------+
           |
           v
+----------------------+
| Check if index empty |
+----------------------+
           |
           v
+----------------------+
| Determine offset to  |
| read (out int32)     |
+----------------------+
           |
           v
+----------------------+
| Calculate byte pos   |
| pos = out * entWidth |
+----------------------+
           |
           v
+----------------------+
| Validate pos + entWidth < i.size |
+----------------------+
           |
           v
+----------------------+
| Read Offset          |
| Read Position        |
+----------------------+
           |
           v
+----------------------+
| Return out, pos, nil  |
+----------------------+
```

**Example: Retrieving Offset `1`**

- **Index Size (`i.size`):** `24 bytes`
- **Entry Width (`entWidth`):** `12 bytes`
- **Input (`in`):** `1`

**Steps:**

1. **Determine Offset:**

   ```go
   out = 1
   ```

2. **Calculate Position:**

   ```go
   pos = 1 * 12 = 12
   ```

3. **Validate Position:**

   ```go
   24 >= 12 + 12 → True
   ```

4. **Read Offset and Position:**

   - **Bytes `12-15` (Offset):** `00 00 00 01` → `1`
   - **Bytes `16-23` (Position):** `00 00 00 00 00 00 00 0D` → `13`

5. **Return:**

   ```go
   out = 1
   pos = 13
   err = nil
   ```

---

## Method: `Write`

```go
func (i *index) Write(off int32, pos uint64) error {
    if uint64(len(i.mmap)) < i.size+entWidth {
        return io.EOF
    }

    enc.PutUint32(i.mmap[i.size:i.size+offWidth], uint32(off))
    enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
    i.size += entWidth
    return nil
}
```

### Purpose

Appends a new index entry to the memory-mapped index file, mapping a given **offset** to its corresponding **position** in the store file.

### Parameters

- **`off int32`**: The offset of the log entry.
- **`pos uint64`**: The byte position in the store file where the log entry begins.

### Returns

- **`error`**: Error object if the write operation fails (e.g., insufficient space).

### Step-by-Step Breakdown

1. **Check for Available Space:**

   ```go
   if uint64(len(i.mmap)) < i.size+entWidth {
       return io.EOF
   }
   ```

   - **Purpose:** Ensure there's enough space in the memory-mapped region to write a new index entry.
   - **Action:** Return `io.EOF` if insufficient space.

2. **Write Offset to Index Entry:**

   ```go
   enc.PutUint32(i.mmap[i.size:i.size+offWidth], uint32(off))
   ```

   - **Purpose:** Encode and write the `offset` as a `uint32` to the current position in the memory-mapped file.
   - **Action:**
     - **Slice:** `i.mmap[i.size : i.size+offWidth]` selects the 4-byte slice.
     - **Encode:** `enc.PutUint32` writes the `offset` in the specified byte order.

3. **Write Position to Index Entry:**

   ```go
   enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos)
   ```

   - **Purpose:** Encode and write the `position` as a `uint64` immediately after the `offset`.
   - **Action:**
     - **Slice:** `i.mmap[i.size+offWidth : i.size+entWidth]` selects the 8-byte slice.
     - **Encode:** `enc.PutUint64` writes the `position` in the specified byte order.

4. **Update Index Size:**

   ```go
   i.size += entWidth
   ```

   - **Purpose:** Increment the current size to account for the newly written index entry.

5. **Return Success:**

   ```go
   return nil
   ```

   - **Action:** Indicate successful write operation.

### Visualization

```
Write Operation:
+----------------------+
| Call Write(off, pos) |
+----------------------+
           |
           v
+----------------------+
| Check available space|
+----------------------+
           |
           v
+----------------------+
| Encode and write off|
| to i.mmap[size:size+4]|
+----------------------+
           |
           v
+----------------------+
| Encode and write pos |
| to i.mmap[size+4:size+12]|
+----------------------+
           |
           v
+----------------------+
| Increment i.size     |
+----------------------+
           |
           v
+----------------------+
| Return nil            |
+----------------------+
```

**Example: Writing Offset `1`, Position `13`**

- **Current Index Size (`i.size`):** `12 bytes`
- **Entry Width (`entWidth`):** `12 bytes`
- **Offset (`off`):** `1`
- **Position (`pos`):** `13`
- **Encoding (`enc`):** `binary.BigEndian`

**Steps:**

1. **Check Space:**

   ```go
   100 >= 12 + 12 → True (Assuming mmap is 100 bytes)
   ```

2. **Write Offset:**

   - **Bytes `12-15`:** `00 00 00 01` → `1`

3. **Write Position:**

   - **Bytes `16-23`:** `00 00 00 00 00 00 00 0D` → `13`

4. **Update Size:**

   ```go
   i.size = 12 + 12 = 24
   ```

5. **Return:**

   ```go
   return nil
   ```

**Visualization:**

```
Before Write:
Memory-Mapped Index (`i.mmap`):
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|00 |00 |00 |00 |00 |00 |00 |00 |00 |00 |00 |00 |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| Offset: 0 | Position: 0 |

After Write:
Memory-Mapped Index (`i.mmap`):
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|00 |00 |00 |00 |00 |00 |00 |00 |00 |00 |00 |00 |
|00 |00 |00 |01 |00 |00 |00 |00 |00 |00 |00 |0D |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
| Offset: 0 | Position: 0 |
| Offset: 1 | Position: 13 |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+
```

---

## Method: `Close`

```go
func (i *index) Close() error {
    // Why both i.mmap.Sync and i.file.Sync?
    if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
        return err
    }

    if err := i.file.Sync(); err != nil {
        return err
    }

    if err := i.file.Truncate(int64(i.size)); err != nil {
        return err
    }

    return i.file.Close()
}
```

### Purpose

Ensures that all changes to the memory-mapped index file are safely persisted to disk before closing the file. This guarantees data integrity and prevents loss of index entries.

### Returns

- **`error`**: Error object if any step during the close operation fails.

### Step-by-Step Breakdown

1. **Sync Memory-Mapped Changes:**

   ```go
   if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
       return err
   }
   ```

   - **Purpose:** Flushes all modifications in the memory-mapped region to the OS's buffer cache.
   - **Flag:** `gommap.MS_SYNC` ensures a synchronous flush.
   - **Action:** Returns an error if the sync fails.

2. **Sync File Descriptor:**

   ```go
   if err := i.file.Sync(); err != nil {
       return err
   }
   ```

   - **Purpose:** Ensures that all buffered data associated with the file descriptor is written to the physical disk.
   - **Action:** Returns an error if the sync fails
