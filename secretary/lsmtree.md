[#04 - Database Storage: Log-Structured Merge Trees & Tuples (CMU Intro to Database Systems)](https://www.youtube.com/watch?v=IHtVWGhG0Xg&t=1372s)

https://github.com/facebook/rocksdb/wiki

https://github.com/krasun/lsmtree
https://github.com/skyzh/mini-lsm




Object stores typically do not use B-trees like databases. Instead, they use hash-based indexing or LSM-trees (Log-Structured Merge Trees) depending on the use case. Here’s why:

1. Hash-Based Indexing (Common for Object Stores)
	•	Object stores like Amazon S3, MinIO, Ceph, and Swift use a distributed key-value store where objects are accessed using a unique key (often a hash of the object name).
	•	Hash maps allow for fast lookups (O(1) complexity) but do not support range queries efficiently.
	•	Metadata (such as object name, size, timestamp) is often stored in a metadata store, which may use databases like PostgreSQL, MySQL, or embedded key-value stores.

2. LSM-Trees (Used in Some Object Stores)
	•	Some object stores use LSM-Trees (Log-Structured Merge Trees) to handle high write throughput efficiently.
	•	Example: RocksDB, LevelDB, which are used in systems like Ceph’s BlueStore.

3. B-Trees (More Common in Databases)
	•	B+ trees are optimized for range queries and ordered data, which is common in databases but not in object stores.
	•	Object stores generally don’t need range queries on keys because they retrieve objects by unique identifiers.

When Might Object Stores Use B-Trees?

Some hybrid systems or metadata layers within object stores might use B-Trees:
	•	Metadata storage: If the metadata layer is implemented in a traditional database (PostgreSQL, MySQL, etc.), it may use B+ trees for indexing.
	•	Filesystems used by Object Stores: If the object store is built on top of a filesystem (e.g., XFS, ext4), those filesystems may use B-trees internally.

Conclusion
	•	B-Trees are great for databases with ordered key access and range queries.
	•	Object stores prioritize fast lookups and scalability, favoring hash maps or LSM-Trees.
	•	Some object store metadata layers might use B-Trees if they rely on traditional databases.

Are you thinking about designing your own object store?




Internals of LSM-Tree vs. B-Tree: A Deep Dive

Both LSM-Trees (Log-Structured Merge-Trees) and B-Trees are used for indexing data, but they are optimized for different workloads. Understanding their internals helps in choosing the right one for a database, key-value store, or file system.

1. Internals of B-Tree

A B-Tree (Balanced Tree) is a self-balancing tree data structure optimized for disk-based storage. It maintains sorted keys and allows efficient point lookups, range queries, insertions, and deletions.

1.1. Structure of a B-Tree
	•	A B-Tree of order d (degree d) has:
	•	A root node (may have fewer keys)
	•	Internal nodes (branching factor of up to 2d)
	•	Leaf nodes (contain actual data or pointers to data)
	•	Each node contains up to 2d keys and 2d+1 child pointers.
	•	Keys in a node are sorted, making binary search within a node possible.

1.2. Operations in B-Tree

1.2.1. Search (O(log N))
	•	Start at the root, perform binary search within the node.
	•	If the key is found, return it.
	•	If not found, follow the correct child pointer and repeat.

1.2.2. Insertion (O(log N))
	1.	Search for the correct leaf node to insert the key.
	2.	If there is space, insert it.
	3.	If the leaf node is full:
	•	Split the node into two.
	•	Push the middle key to the parent.
	•	If the parent is full, repeat the split upwards (recursively).

1.2.3. Deletion (O(log N))
	1.	Find the key in the leaf node.
	2.	If removing it causes an underflow (too few keys), borrow a key from a sibling.
	3.	If borrowing is not possible, merge the node with a sibling.
	4.	If the parent gets underfilled, merge upwards recursively.

1.3. Characteristics of B-Tree
	•	Disk-efficient: Minimizes disk reads by keeping nodes large (typically 4KB, matching disk page sizes).
	•	Well-suited for range queries due to the sorted structure.
	•	Balanced: Ensures O(log N) time complexity for operations.
	•	Mutable: Supports in-place updates without rewriting entire nodes.

2. Internals of LSM-Tree (Log-Structured Merge Tree)

The LSM-Tree is optimized for high write throughput by deferring and batching writes instead of modifying disk structures in-place.

2.1. Structure of LSM-Tree

Instead of modifying data in-place like B-Trees, LSM-Trees follow a write-append strategy with multiple levels of sorted structures.
	1.	MemTable (Memory Table)
	•	An in-memory sorted data structure (usually a Red-Black Tree or Skip List).
	•	Writes go here first.
	•	Fast inserts, but limited in size.
	2.	SSTables (Sorted String Tables) on Disk
	•	When the MemTable fills up, it is flushed to disk as an immutable sorted file (SSTable).
	•	SSTables are sorted and allow efficient range scans.
	3.	Compaction Process
	•	Multiple SSTables are periodically merged (compacted) into larger SSTables, removing old versions of keys.
	•	This reduces read amplification.

2.2. Operations in LSM-Tree

2.2.1. Insertion (O(1) amortized)
	1.	Write to the MemTable (fast, in-memory).
	2.	When the MemTable is full, it is flushed to disk as an SSTable.
	3.	Periodic compaction merges SSTables to optimize read efficiency.

2.2.2. Search (O(log N) or worse due to multiple SSTables)
	1.	Check the MemTable first (fast, in-memory).
	2.	If not found, search recent SSTables on disk.
	3.	If still not found, search older SSTables.
	4.	Bloom filters are used to avoid unnecessary SSTable scans.

2.2.3. Deletion (Tombstones)
	1.	Instead of deleting immediately, a tombstone (delete marker) is written.
	2.	The actual data is removed later during compaction.

2.3. Characteristics of LSM-Tree
	•	Optimized for high write throughput (batching and append-only writes).
	•	Immutable SSTables prevent fragmentation and reduce write amplification.
	•	Compaction reduces read latency but adds extra background work.
	•	Higher read amplification compared to B-Trees (must search multiple SSTables).

3. B-Tree vs. LSM-Tree: Key Differences

Feature	B-Tree	LSM-Tree
Write Speed	Slower (in-place updates, multiple disk I/Os)	Faster (writes to MemTable, append-only SSTables)
Read Speed	Faster (single lookup, O(log N))	Slower (may scan multiple SSTables, higher read amplification)
Range Queries	Efficient (sorted, contiguous leaves)	Less efficient (data spread across SSTables)
Disk Usage	More fragmentation (frequent updates)	More compact (compaction removes old versions)
Compaction Overhead	No compaction needed	Requires background compaction (CPU, I/O overhead)
Write Amplification	Higher (multiple disk writes per update)	Lower (append-only, batch writes)
Read Amplification	Lower (fewer disk reads)	Higher (may read multiple SSTables)
Best for	Balanced workloads (mix of reads & writes)	Write-heavy workloads (logging, key-value stores)

4. When to Use Which?

Use Case	Best Choice
Relational Databases (PostgreSQL, MySQL, etc.)	B-Tree (supports transactions, indexing, and range queries)
Key-Value Stores (RocksDB, LevelDB, etc.)	LSM-Tree (handles high write throughput efficiently)
File Systems (XFS, NTFS, ext4, etc.)	B-Tree (supports random access and metadata storage)
Log Storage (Cassandra, ScyllaDB, HBase, etc.)	LSM-Tree (high write performance and durability)
Distributed Databases (Bigtable, CockroachDB, etc.)	LSM-Tree (optimized for distributed writes and merges)

5. Conclusion
	•	B-Trees are best when reads and writes are balanced, supporting low-latency reads and efficient range scans.
	•	LSM-Trees are best when writes dominate, using batching and compaction to optimize disk usage.
	•	Many modern databases use hybrid approaches, such as B-Trees for metadata and LSM-Trees for logs.

Would you like a deeper dive into a specific area, such as compaction strategies, optimizations, or real-world implementations? 🚀
