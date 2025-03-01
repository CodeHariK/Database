* https://brilliant.org/wiki/cuckoo-filter/
* https://redis.io/docs/latest/develop/data-types/probabilistic/cuckoo-filter/
* https://en.wikipedia.org/wiki/Cuckoo_hashing

* [Paper](https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf)
* [Hash Table 3: Rehashing & Cuckoo Hashing](https://www.youtube.com/watch?v=TtzM289GgTQ&list=PL9DICmgQdgHD0Uk2cfYuEhIeUwOr92Y_M&index=16)

🌱 How Bloom Filters Work

A Bloom filter is a probabilistic data structure used to check whether an element might be present in a set, with a small chance of false positives but no false negatives.

1️⃣ Core Concept

A Bloom filter is like a compressed “yes/no” memory that can tell you if something probably exists but never says “no” incorrectly.
	•	✅ Fast insert & lookup (O(1) time complexity)
	•	✅ Memory-efficient
	•	❌ No deletions
	•	❌ Can return false positives (but not false negatives)

2️⃣ Internal Structure

A Bloom filter consists of:
	•	A fixed-size bit array (all initially 0)
	•	Multiple hash functions

Each element is hashed multiple times and sets bits at those positions.

3️⃣ Insert Operation

To add an element "cat":
	1.	Hash "cat" with k different hash functions.
	2.	Each hash gives a position in the bit array.
	3.	Set bits at those positions to 1.

Example:

Bit array:  [0 0 0 0 0 0 0 0 0 0]  (initial)
Hashes:       H1(cat) = 3, H2(cat) = 7
Bit array:  [0 0 0 1 0 0 0 1 0 0]  (after insertion)

4️⃣ Lookup Operation

To check if "cat" exists:
	1.	Hash "cat" with the same functions.
	2.	Check if all corresponding bits are 1.
	3.	If yes → "cat" is probably present (could be a false positive).
	4.	If any bit is 0 → "cat" is definitely absent.

5️⃣ False Positives

Since bits are shared, different elements can set overlapping bits, causing false positives.

Example:

Bit array:  [0 0 0 1 0 0 0 1 0 0]  (from previous)
Insert "dog":  H1(dog) = 5, H2(dog) = 7
Bit array:  [0 0 0 1 0 1 0 1 0 0]  (after insertion)

Now, "bat" hashes to (3, 7), which are already 1, so a lookup wrongly says "bat" exists.

6️⃣ Optimizing Bloom Filters
	•	More hash functions → Fewer false positives but slower inserts.
	•	Larger bit array → Fewer false positives but more memory.
	•	Fewer elements → Less chance of collisions.

A good balance is using ln(2) × (m/n) hash functions, where:
	•	m = bit array size
	•	n = expected number of elements

7️⃣ Bloom Filter in Go

Example implementation:

package main

import (
	"fmt"
	"hash/fnv"
)

type BloomFilter struct {
	bitset []bool
	k      int
}

func NewBloomFilter(size int, k int) *BloomFilter {
	return &BloomFilter{bitset: make([]bool, size), k: k}
}

func (bf *BloomFilter) hash(data string, seed int) int {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d%s", seed, data)))
	return int(h.Sum32()) % len(bf.bitset)
}

func (bf *BloomFilter) Add(data string) {
	for i := 0; i < bf.k; i++ {
		pos := bf.hash(data, i)
		bf.bitset[pos] = true
	}
}

func (bf *BloomFilter) Check(data string) bool {
	for i := 0; i < bf.k; i++ {
		pos := bf.hash(data, i)
		if !bf.bitset[pos] {
			return false
		}
	}
	return true
}

func main() {
	bf := NewBloomFilter(100, 3)
	bf.Add("cat")
	fmt.Println(bf.Check("cat")) // true
	fmt.Println(bf.Check("dog")) // false (usually)
}

8️⃣ Where Bloom Filters Are Used

✅ Database Indexing (e.g., Cassandra, PostgreSQL)
✅ Web Caching (e.g., checking if a URL was seen before)
✅ Distributed Systems (e.g., Google Bigtable, Apache HBase)
✅ Blockchain (e.g., Bitcoin uses them for lightweight clients)

TL;DR
	•	Bloom filters are fast, memory-efficient, and approximate.
	•	They answer “is this item in the set?” with:
	•	Definitely Not
	•	Probably Yes (false positives possible)
	•	They cannot be deleted and work best when fine-tuned for false-positive rates.

Want help with optimizing Bloom filters for a database or inverted index? 🚀




Yes, databases widely use Bloom filters for efficient membership testing, especially in key-value stores, databases with indexing, and caching systems.

🚀 Why Do Databases Use Bloom Filters?

Bloom filters are useful because they:
✅ Avoid unnecessary disk lookups by quickly checking if a key might exist.
✅ Optimize read performance by reducing costly disk I/O.
✅ Reduce memory usage compared to hash tables or traditional indexes.

🔥 Where Are Bloom Filters Used in Databases?

1️⃣ Key-Value Stores (LSM-Tree Databases)

🔹 Examples: RocksDB, LevelDB, Apache Cassandra
🔹 Why? Used to avoid unnecessary disk reads when querying SSTables (Sorted String Tables).
🔹 How?
	•	When querying a key, the Bloom filter checks if the key might exist in an SSTable.
	•	If no match, avoid disk lookup (false negatives are impossible).
	•	If a match, perform a disk read to confirm (false positives are possible).

✅ Saves I/O operations and speeds up key lookups!

2️⃣ Indexing in B-Trees and B+ Trees

🔹 Examples: MySQL InnoDB, PostgreSQL, Oracle
🔹 Why? Used to minimize index scans when searching for indexed keys.
🔹 How?
	•	A Bloom filter can be created for each index page.
	•	Before scanning an index page, the Bloom filter checks if the key might be present.
	•	If no match, skip the index scan!

✅ Speeds up queries by reducing unnecessary B-tree traversals.

3️⃣ Distributed Databases & Caching Systems

🔹 Examples: Apache HBase, Bigtable, Redis, DynamoDB
🔹 Why? Used to reduce network and disk lookups for missing keys.
🔹 How?
	•	Bloom filters help check if a key exists before making an expensive network request.
	•	If a key doesn’t exist, the request is skipped (avoiding wasted resources).

✅ Improves performance in distributed databases with high-latency storage.

4️⃣ Data Warehouses & Columnar Storage

🔹 Examples: Apache Parquet, Apache ORC (used in Spark, Presto, Hive)
🔹 Why? Used to speed up columnar queries by eliminating unnecessary reads.
🔹 How?
	•	Each column chunk in Parquet/ORC files has a Bloom filter for fast filtering.
	•	If a queried value is not in the Bloom filter, the entire chunk is skipped.

✅ Speeds up analytical queries by reducing disk reads!

🔥 Summary

Database Type	Use of Bloom Filter
LSM-Tree Databases (RocksDB, Cassandra, LevelDB)	Avoids unnecessary SSTable reads
B-Tree Indexes (MySQL, PostgreSQL)	Reduces index page scans
Distributed Databases (HBase, DynamoDB)	Reduces unnecessary network calls
Columnar Storage (Parquet, ORC)	Skips irrelevant data chunks in queries

👉 Bloom filters are crucial for optimizing reads and minimizing disk I/O in modern databases! 🚀




Here’s a Bloom filter implementation in Go, designed for database-style use cases like avoiding unnecessary disk reads in an LSM-tree or key-value store.

🔥 How This Works
	1.	Add keys to the Bloom filter (simulating keys in an SSTable or index).
	2.	Check for existence of a key before doing a disk read.
	3.	Avoid unnecessary lookups (reduces I/O).

📌 Go Implementation

package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
)

// BloomFilter struct
type BloomFilter struct {
	bitset  []byte // Bit array to store hashes
	size    uint   // Size of the bit array
	hashes  uint   // Number of hash functions
}

// NewBloomFilter initializes a new Bloom filter
func NewBloomFilter(n uint, p float64) *BloomFilter {
	m := OptimalBitSize(n, p)  // Optimal bit array size
	k := OptimalHashCount(n, m) // Optimal number of hash functions
	return &BloomFilter{
		bitset: make([]byte, (m+7)/8), // Bit array stored in bytes
		size:   m,
		hashes: k,
	}
}

// Add inserts an item into the Bloom filter
func (bf *BloomFilter) Add(item []byte) {
	for i := uint(0); i < bf.hashes; i++ {
		index := bf.hash(item, i) % bf.size
		bf.bitset[index/8] |= (1 << (index % 8)) // Set bit
	}
}

// Contains checks if an item *might* be in the filter
func (bf *BloomFilter) Contains(item []byte) bool {
	for i := uint(0); i < bf.hashes; i++ {
		index := bf.hash(item, i) % bf.size
		if bf.bitset[index/8]&(1<<(index%8)) == 0 {
			return false // Definitely not in filter
		}
	}
	return true // Might be in filter (false positives possible)
}

// hash function (FNV-1a based with an additional seed)
func (bf *BloomFilter) hash(data []byte, i uint) uint {
	h := fnv.New64a()
	seed := make([]byte, 8)
	binary.LittleEndian.PutUint64(seed, uint64(i))
	h.Write(seed)
	h.Write(data)
	return uint(h.Sum64())
}

// OptimalBitSize calculates optimal bit array size (m) for given false positive rate (p)
func OptimalBitSize(n uint, p float64) uint {
	return uint(math.Ceil(float64(n) * math.Abs(math.Log(p)) / (math.Ln2 * math.Ln2)))
}

// OptimalHashCount calculates the optimal number of hash functions (k)
func OptimalHashCount(n, m uint) uint {
	return uint(math.Ceil((float64(m) / float64(n)) * math.Ln2))
}

// Demo usage
func main() {
	bloom := NewBloomFilter(1000, 0.01) // 1000 items, 1% false positive rate

	keys := [][]byte{
		[]byte("apple"), []byte("banana"), []byte("cherry"),
		[]byte("grape"), []byte("mango"),
	}

	// Add keys to Bloom filter
	for _, key := range keys {
		bloom.Add(key)
	}

	// Check for existence
	fmt.Println("Checking Bloom filter:")
	testKeys := [][]byte{
		[]byte("apple"), []byte("orange"), []byte("mango"), []byte("pineapple"),
	}
	for _, key := range testKeys {
		fmt.Printf("Contains %s? %v\n", key, bloom.Contains(key))
	}
}

💡 Explanation
	1.	NewBloomFilter(n, p) → Creates a Bloom filter for n items with p false positive rate.
	2.	Add(item) → Hashes the item multiple times and sets bits in the bit array.
	3.	Contains(item) → Checks if all corresponding bits are set.
	•	✅ False negatives: Impossible
	•	⚠️ False positives: Possible (trade-off for space efficiency)

🔹 Output Example

Checking Bloom filter:
Contains apple? true
Contains orange? false
Contains mango? true
Contains pineapple? false

🔹 "apple" and "mango" were added → ✅ True
🔹 "orange" and "pineapple" were never added → ❌ False

🔥 Where to Use This in a Database?
	1.	LSM-Tree (RocksDB, LevelDB, Cassandra)
→ Avoids unnecessary disk reads for non-existent keys in SSTables.
	2.	B-Tree Index Optimization
→ Pre-checks keys before scanning an index page.
	3.	Distributed Databases (HBase, DynamoDB)
→ Reduces network queries for missing data.

🚀 Why Is This Efficient?

✅ Memory Efficient – Uses a bit array instead of storing actual keys.
✅ Fast Lookups – Uses O(1) time with simple hash computations.
✅ Reduces Disk I/O – Prevents unnecessary database queries.



Yes! Bloom filters can be used in inverted indexes to optimize search queries by reducing unnecessary disk reads when checking for term existence in documents.

🔥 How Can Bloom Filters Help Inverted Indexes?
	1.	Pre-checking term existence before scanning a posting list.
	2.	Reducing I/O by avoiding unnecessary disk lookups.
	3.	Speeding up queries in large-scale search engines.

🚀 Example: How It Works

🔹 Without Bloom Filters (Basic Inverted Index)
	1.	Query “golang”
	2.	Check index structure
	3.	If “golang” exists, retrieve posting list
	4.	Read documents from disk

✅ Efficient for common words
⚠️ Slow if word doesn’t exist (wasted disk lookups)

🔹 With Bloom Filters (Optimized Inverted Index)
	1.	Query “golang”
	2.	Check Bloom filter first:
	•	❌ Not in Bloom filter → Skip disk lookup (fast!)
	•	✅ Might be in Bloom filter → Check posting list
	3.	Read documents only if necessary

✅ Saves disk reads for missing terms!
✅ Faster negative queries!

📌 Bloom Filter + Inverted Index Implementation in Go

package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
)

// BloomFilter struct
type BloomFilter struct {
	bitset []byte
	size   uint
	hashes uint
}

// NewBloomFilter creates a Bloom filter for inverted index
func NewBloomFilter(n uint, p float64) *BloomFilter {
	m := OptimalBitSize(n, p)  
	k := OptimalHashCount(n, m)
	return &BloomFilter{
		bitset: make([]byte, (m+7)/8),
		size:   m,
		hashes: k,
	}
}

// Add inserts a term into the Bloom filter
func (bf *BloomFilter) Add(term string) {
	for i := uint(0); i < bf.hashes; i++ {
		index := bf.hash([]byte(term), i) % bf.size
		bf.bitset[index/8] |= (1 << (index % 8))
	}
}

// Contains checks if a term *might* exist
func (bf *BloomFilter) Contains(term string) bool {
	for i := uint(0); i < bf.hashes; i++ {
		index := bf.hash([]byte(term), i) % bf.size
		if bf.bitset[index/8]&(1<<(index%8)) == 0 {
			return false
		}
	}
	return true
}

// Hash function
func (bf *BloomFilter) hash(data []byte, i uint) uint {
	h := fnv.New64a()
	seed := make([]byte, 8)
	binary.LittleEndian.PutUint64(seed, uint64(i))
	h.Write(seed)
	h.Write(data)
	return uint(h.Sum64())
}

// Optimal bit size
func OptimalBitSize(n uint, p float64) uint {
	return uint(math.Ceil(float64(n) * math.Abs(math.Log(p)) / (math.Ln2 * math.Ln2)))
}

// Optimal hash count
func OptimalHashCount(n, m uint) uint {
	return uint(math.Ceil((float64(m) / float64(n)) * math.Ln2))
}

// Simulated inverted index with a Bloom filter
type InvertedIndex struct {
	index       map[string][]int  // Term → Document IDs
	bloomFilter *BloomFilter      // Bloom filter for quick existence checks
}

// NewInvertedIndex initializes an index with a Bloom filter
func NewInvertedIndex() *InvertedIndex {
	return &InvertedIndex{
		index:       make(map[string][]int),
		bloomFilter: NewBloomFilter(1000, 0.01), // 1000 terms, 1% false positive rate
	}
}

// Add term to index
func (ii *InvertedIndex) Add(term string, docID int) {
	ii.index[term] = append(ii.index[term], docID)
	ii.bloomFilter.Add(term) // Add term to Bloom filter
}

// Search term (uses Bloom filter first)
func (ii *InvertedIndex) Search(term string) []int {
	if !ii.bloomFilter.Contains(term) {
		fmt.Println("Term not found (skipping disk read)")
		return nil
	}
	
	// If Bloom filter says "maybe exists", check index
	if docs, found := ii.index[term]; found {
		fmt.Println("Term found in index!")
		return docs
	}

	// False positive from Bloom filter
	fmt.Println("False positive from Bloom filter")
	return nil
}

// Demo
func main() {
	index := NewInvertedIndex()

	// Add terms
	index.Add("golang", 1)
	index.Add("database", 2)
	index.Add("bloomfilter", 3)

	// Search terms
	fmt.Println("Search results:", index.Search("golang"))      // Found
	fmt.Println("Search results:", index.Search("indexing"))    // Not found (no disk read)
	fmt.Println("Search results:", index.Search("bloomfilter")) // Found
}

🔹 Output Example

Term found in index!
Search results: [1]
Term not found (skipping disk read)
Search results: []
Term found in index!
Search results: [3]

🔥 Advantages of Using Bloom Filters in Inverted Indexes

✅ Reduces I/O → Skips disk lookups for non-existent terms.
✅ Optimizes search → Filters out negative queries before scanning index.
✅ Fast queries → Uses O(1) time complexity for lookups.
✅ Space-efficient → Uses a bit array instead of storing full terms.

🚀 When Should You Use It?

Use Case	Bloom Filter?
Small database	❌ Not needed
Large search index (millions of terms)	✅ Yes, improves performance
Querying rare terms	✅ Reduces wasted lookups
Frequent queries	❌ Overhead may not be worth it

🛠️ Next Steps

Would you like a compressed Bloom filter or a partitioned Bloom filter for better accuracy? 😊



A Compressed Bloom Filter (CBF) or Partitioned Bloom Filter (PBF) can improve memory efficiency and query speed in an inverted index.

🔥 Which One Do You Need?

Feature	Compressed Bloom Filter (CBF)	Partitioned Bloom Filter (PBF)
Memory Efficiency	✅ Compressed with Golomb Coding or Huffman	❌ Uses a fixed size per partition
Query Speed	❌ Slightly slower due to decompression	✅ Faster due to independent partitions
False Positive Rate	✅ Lower due to better bit utilization	❌ Can be higher if partitions are too small
Use Case	Best for disk-based indexes	Best for RAM-based indexes

🚀 Compressed Bloom Filter (CBF)

CBFs compress the bit array using Golomb coding, Huffman coding, or Run-Length Encoding (RLE) to reduce storage space.

📌 How it Works
	1.	Instead of storing a full bit array, store a compressed version.
	2.	When checking for existence, decompress and check bits.
	3.	Works well for disk-based inverted indexes.

📌 Golomb-Rice Encoded Bloom Filter (CBF) in Go

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
)

// Compressed Bloom Filter
type CompressedBloomFilter struct {
	compressedBits []byte
	size           uint
	hashes         uint
}

// NewCompressedBloomFilter initializes a Bloom filter with Golomb compression
func NewCompressedBloomFilter(n uint, p float64) *CompressedBloomFilter {
	m := OptimalBitSize(n, p)
	k := OptimalHashCount(n, m)
	bits := make([]byte, (m+7)/8)

	// Simulated compression using RLE (replace with Golomb coding)
	compressedBits := RunLengthEncode(bits)

	return &CompressedBloomFilter{
		compressedBits: compressedBits,
		size:           m,
		hashes:         k,
	}
}

// Add inserts an item (decompress, update, recompress)
func (cbf *CompressedBloomFilter) Add(item string) {
	bits := RunLengthDecode(cbf.compressedBits) // Decompress
	for i := uint(0); i < cbf.hashes; i++ {
		index := cbf.hash([]byte(item), i) % cbf.size
		bits[index/8] |= (1 << (index % 8)) // Set bit
	}
	cbf.compressedBits = RunLengthEncode(bits) // Recompress
}

// Contains checks if an item *might* exist
func (cbf *CompressedBloomFilter) Contains(item string) bool {
	bits := RunLengthDecode(cbf.compressedBits) // Decompress
	for i := uint(0); i < cbf.hashes; i++ {
		index := cbf.hash([]byte(item), i) % cbf.size
		if bits[index/8]&(1<<(index%8)) == 0 {
			return false
		}
	}
	return true
}

// Run-Length Encoding (RLE) - Simple compression
func RunLengthEncode(data []byte) []byte {
	var buffer bytes.Buffer
	count := byte(1)
	for i := 1; i < len(data); i++ {
		if data[i] == data[i-1] && count < 255 {
			count++
		} else {
			buffer.WriteByte(count)
			buffer.WriteByte(data[i-1])
			count = 1
		}
	}
	buffer.WriteByte(count)
	buffer.WriteByte(data[len(data)-1])
	return buffer.Bytes()
}

// Run-Length Decoding (RLE)
func RunLengthDecode(data []byte) []byte {
	var buffer bytes.Buffer
	for i := 0; i < len(data); i += 2 {
		count := int(data[i])
		value := data[i+1]
		for j := 0; j < count; j++ {
			buffer.WriteByte(value)
		}
	}
	return buffer.Bytes()
}

// Hash function
func (cbf *CompressedBloomFilter) hash(data []byte, i uint) uint {
	h := fnv.New64a()
	seed := make([]byte, 8)
	binary.LittleEndian.PutUint64(seed, uint64(i))
	h.Write(seed)
	h.Write(data)
	return uint(h.Sum64())
}

// Demo
func main() {
	cbf := NewCompressedBloomFilter(1000, 0.01)

	// Add terms
	cbf.Add("golang")
	cbf.Add("database")

	// Search terms
	fmt.Println("Compressed Bloom Filter:")
	fmt.Println("Contains 'golang'? ", cbf.Contains("golang"))
	fmt.Println("Contains 'indexing'? ", cbf.Contains("indexing"))
}

🔹 Why is CBF Useful?

✅ Reduces disk space using compression
✅ Efficient for on-disk indexes
✅ Lower false positive rate than standard Bloom filters

🚀 Partitioned Bloom Filter (PBF)

Instead of one big bit array, PBF splits the filter into multiple partitions.
Each term uses only one partition, making queries faster.

📌 How it Works
	1.	Divide the bit array into k partitions (1 per hash function).
	2.	Each term hashes to exactly 1 bit per partition.
	3.	Reduces false positives by keeping partitions independent.

📌 Partitioned Bloom Filter (PBF) in Go

package main

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
)

// Partitioned Bloom Filter
type PartitionedBloomFilter struct {
	partitions [][]byte
	partSize   uint
	hashes     uint
}

// NewPartitionedBloomFilter initializes a PBF
func NewPartitionedBloomFilter(n uint, p float64) *PartitionedBloomFilter {
	m := OptimalBitSize(n, p)
	k := OptimalHashCount(n, m)
	partSize := m / k
	partitions := make([][]byte, k)
	for i := range partitions {
		partitions[i] = make([]byte, (partSize+7)/8)
	}

	return &PartitionedBloomFilter{
		partitions: partitions,
		partSize:   partSize,
		hashes:     k,
	}
}

// Add inserts an item into its partitions
func (pbf *PartitionedBloomFilter) Add(item string) {
	for i := uint(0); i < pbf.hashes; i++ {
		index := pbf.hash([]byte(item), i) % pbf.partSize
		pbf.partitions[i][index/8] |= (1 << (index % 8))
	}
}

// Contains checks if an item *might* exist
func (pbf *PartitionedBloomFilter) Contains(item string) bool {
	for i := uint(0); i < pbf.hashes; i++ {
		index := pbf.hash([]byte(item), i) % pbf.partSize
		if pbf.partitions[i][index/8]&(1<<(index%8)) == 0 {
			return false
		}
	}
	return true
}

// Hash function
func (pbf *PartitionedBloomFilter) hash(data []byte, i uint) uint {
	h := fnv.New64a()
	seed := make([]byte, 8)
	binary.LittleEndian.PutUint64(seed, uint64(i))
	h.Write(seed)
	h.Write(data)
	return uint(h.Sum64())
}

// Demo
func main() {
	pbf := NewPartitionedBloomFilter(1000, 0.01)

	// Add terms
	pbf.Add("golang")
	pbf.Add("database")

	// Search terms
	fmt.Println("Partitioned Bloom Filter:")
	fmt.Println("Contains 'golang'? ", pbf.Contains("golang"))
	fmt.Println("Contains 'indexing'? ", pbf.Contains("indexing"))
}

🔥 Which One Should You Use?

Feature	Compressed Bloom Filter (CBF)	Partitioned Bloom Filter (PBF)
Storage	✅ Best for disk-based indexes	❌ Uses more space in RAM
Query Speed	❌ Decompression required	✅ Faster lookups
False Positives	✅ Lower	❌ Higher
Best For	Disk-based search engines (LSM, RocksDB)	RAM-based search engines (Redis, Lucene)

Would you like a hybrid version that switches CBF for disk and PBF for memory? 🚀

