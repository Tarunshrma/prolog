Append Process Documentation
Overview
When appending data to a store, especially in systems like logs or queues, it's common to prefix each data entry with its length. This approach allows for efficient reading and parsing later on.

Table of Contents

Append Process Overview
Append Method Steps
Visual Example
Initial State
First Append: "Hello"
Step-by-Step Process
State After First Append
Second Append: "World!"
Step-by-Step Process
State After Second Append
Buffer Flush
State After Flushing
Conclusion
Append Process Overview
(This section can be expanded with additional details if needed.)

Append Method Steps
Here's a high-level overview of the steps involved in the Append method:

Lock the Store: Ensure thread-safe access.
Record the Current Position: Note where the new data will start.
Write the Length of the Data: Encode and write the length as a fixed-size header.
Write the Actual Data: Append the data itself.
Update the Store Size: Reflect the new total size.
Unlock the Store: Release the lock.
Visual Example
Let's walk through an example where we append two pieces of data to the store: "Hello" and "World!". We'll visualize how each append operation affects the store's buffer and the underlying file.

Initial State
Store Size (s.size): 0 bytes
Buffer (s.buf): Empty
File (file): Empty
plaintext
Copy code
Store:
+---------+-----+-----+-----+-----+-----+-----+-----+-----+
|  Size: 0 bytes                                      ... |
+---------+-----+-----+-----+-----+-----+-----+-----+-----+

Buffer:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|                                             (empty)    |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+

File:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|                                             (empty)    |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
First Append: "Hello"
Data to Append (p): "Hello"
Length (len(p)): 5 bytes
Length Encoding: uint64 (8 bytes)
Encoding Order: BigEndian
Step-by-Step Process
Lock the Store: Ensure exclusive access.
Record Position: pos = 0 (current s.size)
Write Length:
Convert 5 to uint64 in BigEndian:
plaintext
Copy code
00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000101
Hex Representation: 00 00 00 00 00 00 00 05
Write these 8 bytes to the buffer using binary.Write.
Write Data: Write "Hello" (5 bytes) to the buffer using s.buf.Write.
Update Size: s.size += 8 (lenWidth) + 5 (data) = 13 bytes
Unlock the Store: Release the lock.
State After First Append
plaintext
Copy code
Store:
+---------+-----+-----+-----+-----+-----+-----+-----+-----+
|  Size: 13 bytes                                     ... |
+---------+-----+-----+-----+-----+-----+-----+-----+-----+

Buffer:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|00 00 00 00 00 00 00 05|H|e|l|l|o|               |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+

File:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|                                             (empty)    |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
Explanation:

binary.Write writes the length (5) as 00 00 00 00 00 00 00 05 to the buffer.
s.buf.Write appends the actual data "Hello" to the buffer.
The buffer now holds both the length and the data but hasn't yet flushed them to the file.
Second Append: "World!"
Data to Append (p): "World!"
Length (len(p)): 6 bytes
Length Encoding: uint64 (8 bytes)
Encoding Order: BigEndian
Step-by-Step Process
Lock the Store: Ensure exclusive access.
Record Position: pos = 13 (current s.size)
Write Length:
Convert 6 to uint64 in BigEndian:
plaintext
Copy code
00000000 00000000 00000000 00000000 00000000 00000000 00000000 00000110
Hex Representation: 00 00 00 00 00 00 00 06
Write these 8 bytes to the buffer using binary.Write.
Write Data: Write "World!" (6 bytes) to the buffer using s.buf.Write.
Update Size: s.size += 8 (lenWidth) + 6 (data) = 27 bytes
Unlock the Store: Release the lock.
State After Second Append
plaintext
Copy code
Store:
+---------+-----+-----+-----+-----+-----+-----+-----+-----+
|  Size: 27 bytes                                     ... |
+---------+-----+-----+-----+-----+-----+-----+-----+-----+

Buffer:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|00 00 00 00 00 00 00 05|H|e|l|l|o|00 00 00 00 00 00 00 06|W|o|r|l|d|!|
+-----+-----+-----+-----+-----+-----+-----+-----+-----+

File:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|                                             (empty)    |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
Explanation:

binary.Write writes the length (6) as 00 00 00 00 00 00 00 06 to the buffer.
s.buf.Write appends the actual data "World!" to the buffer.
The buffer now contains two entries:
First Entry: 00 00 00 00 00 00 00 05 | H | e | l | l | o
Second Entry: 00 00 00 00 00 00 00 06 | W | o | r | l | d | !
The buffer holds all data, pending a flush to the file.
Buffer Flush
At some point, either automatically when the buffer is full or manually by calling s.buf.Flush(), the buffered data is written to the underlying file.

State After Flushing
plaintext
Copy code
Store:
+---------+-----+-----+-----+-----+-----+-----+-----+-----+
|  Size: 27 bytes                                     ... |
+---------+-----+-----+-----+-----+-----+-----+-----+-----+

Buffer:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|                                             (empty)    |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+

File:
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
|00 00 00 00 00 00 00 05|H|e|l|l|o|00 00 00 00 00 00 00 06|W|o|r|l|d|!|
+-----+-----+-----+-----+-----+-----+-----+-----+-----+
Explanation:

Buffer is Flushed: All buffered data is written to the file.
File Now Contains:
First Entry: 00 00 00 00 00 00 00 05 | H | e | l | l | o
Second Entry: 00 00 00 00 00 00 00 06 | W | o | r | l | d | !
Buffer is Empty: Ready to accept new data.

Conclusion

This documentation outlines the append process used to efficiently store and manage data entries with prefixed lengths, ensuring thread-safe operations and optimized read/write performance. By following the step-by-step process and understanding the state transitions, developers can implement and maintain robust data storage mechanisms in their systems.