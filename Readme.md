#

* [#08 - Tree Indexes: B+Trees (CMU Intro to Database Systems)](https://www.youtube.com/watch?v=scUtG_6M_lU)
* [SQLite: How it works](https://www.youtube.com/watch?v=ZSKLA81tBis)
* [Write a database from scratch](https://www.youtube.com/playlist?list=PLWRwj01AnyEtjaw-ZnnAQWnVYPZF5WayV)

* [Understanding B-Trees: The Data Structure Behind Modern Databases](https://www.youtube.com/watch?v=K1a2Bk8NrYQ)

* [Build a NoSQL Database From Scratch in 1000 Lines of Code](https://medium.com/better-programming/build-a-nosql-database-from-the-scratch-in-1000-lines-of-code-8ed1c15ed924)
* [Writing a SQL database from scratch in Go: 1. SELECT, INSERT, CREATE and a REPL](https://notes.eatonphil.com/database-basics.html)

##

```
1. Persistence. How not to lose or corrupt your data. Recovering from a crash.
2. Indexing. Efficiently querying and manipulating your data. (B-tree).
3. Concurrency. How to handle multiple (large number of ) clients. And transactions.
```

### Persistence
Why do we need databases? Why not dump the data directly into files

Let’s say your process crashed middle-way while writing to a file, or you lost power, what’s
the state of the file?
• Does the file just lose the last write?
• Or ends up with a half-written file?
• Or ends up in an even more corrupted state?
Any outcome is possible. Your data is not guaranteed to persist on a disk when you simply
write to files. This is a concern of databases. And a database will recover to a usable state
when started after an unexpected shutdown.
Can we achieve persistence without using a database? There is a way:
1. Write the whole updated dataset to a new file.
2. Call fsync on the new file.
3. Overwrite the old file by renaming the new file to the old file, which is guaranteed
by the file systems to be atomic.
This is only acceptable when the dataset is tiny. A database like SQLite can do incremental
updates.

### Indexing

• Analytical (OLAP) queries typically involve a large amount of data, with aggregation,
grouping, or join operations.
• In contrast, transactional (OLTP) queries usually only touch a small amount of
indexed data. The most common types of queries are indexed point queries and
indexed range queries.

Data structures that persist on a disk to look
up data are called “indexes” in database systems. And database indexes can be larger than
memory. There is a saying: if your problem fits in memory, it’s an easy problem.
Common data structures for indexing include B-Trees and LSM-Trees.

1. Scan the whole data set. (No index is used).
2. Point query: Query the index by a specific key.
3. Range query: Query the index by a range. (The index is sorted).

#### Data structure

On-disk data structures are often used when the amounts of data are so large that
keeping an entire dataset in memory is impossible or not feasible. Only a fraction of
the data can be cached in memory at any time, and the rest has to be stored on disk in
a manner that allows efficiently accessing it.

On spinning disks, seeks increase costs of random reads because they require disk
rotation and mechanical head movements to position the read/write head to the
desired location. However, once the expensive part is done, reading or writing contig‐
uous bytes (i.e., sequential operations) is relatively cheap.
The smallest transfer unit of a spinning drive is a sector, so when some operation is
performed, at least an entire sector can be read or written. Sector sizes typically range
from 512 bytes to 4 Kb.
Head positioning is the most expensive part of an operation on the HDD. This is one
of the reasons we often hear about the positive effects of sequential I/O: reading and
writing contiguous memory segments from disk.

In SSDs, we don’t have a strong emphasis on random versus sequential I/O, as in
HDDs, because the difference in latencies between random and sequential reads is
not as large. There is still some difference caused by prefetching, reading contiguous
pages, and internal parallelism

Writing only full blocks, and combining subsequent writes to the same block, can
help to reduce the number of required I/O operations.

In summary, on-disk structures are designed with their target storage specifics in
mind and generally optimize for fewer disk accesses. We can do this by improving
locality, optimizing the internal representation of the structure, and reducing the
number of out-of-page pointers.

##### Hashtable 
no sorting or ordering, resizing problems

##### Binary search tree
Unbalanced trees have a worst-case complexity of O(N).
Balanced trees give us an average O(log2 N). At the same time, due to low fanout
(fanout is the maximum allowed number of children per node), we have to perform
balancing, relocate nodes, and update pointers rather frequently. Increased mainte‐
nance costs make BSTs impractical as on-disk data structures

If we wanted to maintain a BST on disk, we’d face several problems. One problem is
locality: since elements are added in random order, there’s no guarantee that a newly
created node is written close to its parent, which means that node child pointers may
span across several disk pages. We can improve the situation to a certain extent by
modifying the tree layout and using paged binary trees 

Another problem, closely related to the cost of following child pointers, is tree height.
Since binary trees have a fanout of just two, height is a binary logarithm of the num‐
ber of the elements in the tree, and we have to perform O(log2 N) seeks to locate the
searched element and, subsequently, perform the same number of disk transfers. 2-3-
Trees and other low-fanout trees have a similar limitation: while they are useful as
in-memory data structures, small node size makes them impractical for external storage

A naive on-disk BST implementation would require as many disk seeks as compari‐
sons, since there’s no built-in concept of locality. 

Considering these factors, a version of the tree that would be better suited for disk
implementation has to exhibit the following properties:
• High fanout to improve locality of the neighboring keys.
• Low height to reduce the number of seeks during traversal.

##### Balanced binary trees BTree 
Queried and updated in O(log(n)) and can be range-queried. A BTree is roughly a balanced n-ary tree
Why use an n-ary tree instead of a binary tree => 
1. Less space overhead
Every leaf node in a binary tree is reached via a pointer from a parent node, and
the parent node may also have a parent. On average, each leaf node requires 1~2
pointers.
This is in contrast to B-trees, where multiple data in a leaf node share one parent.
And n-ary trees are also shorter. Less space is wasted on pointers.
2. Faster in memory.
Due to modern CPU memory caching and other factors, n-ary trees can be faster
than binary trees, even if their big-O complexity is the same.
3. Less disk IO.
• B-trees are shorter, which means fewer disk seeks.
• The minimum size of disk IOs is usually the size of the memory page (probably
4K). The operating system will fill the whole 4K page even if you read a smaller
size. It’s optimal if we make use of all the information in a 4K page (by choosing
the node size of at least one page).

##### Log-structured merge-tree LSM-Trees
How to query:
1. An LSM-Tree contains multiple levels of data.
2. Each level is sorted and split into multiple files.
3. A point query starts at the top level, if the key is not found, the search continues to
the next level.
4. A range query merges the results from all levels, higher levels have more priority
when merging.
How to update:
5. When updating a key, the key is inserted into a file from the top level first.
6. If the file size exceeds a threshold, merge it with the next level.
7. The file size threshold increases exponentially with each level, which means that
the amount of data also increases exponentially.
Let’s analyze how this works. For queries:
1. Each level is sorted, keys can be found via binary search, and range queries are just
sequential file IO. It’s efficient.
For updates:
2. The top-level file size is small, so inserting into the top level requires only a small
amount of IO.
3. Data is eventually merged to a lower level. Merging is sequential IO, which is an
advantage.
4. Higher levels trigger merging more often, but the merge is also smaller.
5. When merging a file into a lower level, any lower files whose range intersects are
replaced by the merged results (which can be multiple files). We can see why levels
are split into multiple files — to reduce the size of the merge.
6. Merging can be done in the background. However, low-level merging can suddenly
cause high IO usage, which can degrade system performance.
