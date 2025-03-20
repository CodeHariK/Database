package secretary

import (
	"fmt"
	"math"
	"os"

	"github.com/dgraph-io/ristretto/v2"
)

/*

https://github.com/sqlite/sqlite/blob/master/src/pager.c

What is a Pager?

A pager is a low-level component responsible for reading and writing fixed-size pages to and from storage (disk, SSD, or memory).
It acts as an abstraction layer between the storage system and higher-level database structures.

Responsibilities of a Pager:
1.	Reading Pages: When a page is requested, the pager loads it from disk (if not already in memory).
2.	Writing Pages: When pages are modified, the pager ensures they are written back to disk properly.
3.	Page Allocation & Freeing: It manages free pages and allocates new pages as needed.
4.	Crash Recovery: Works with journaling or WAL (Write-Ahead Logging) to ensure data consistency.
5.	Interacting with the Buffer Pool: The pager fetches pages into the buffer pool and evicts them when necessary.

A buffer pool is a memory management component in a database system that caches frequently accessed data pages in RAM. It helps reduce disk I/O by keeping recently used or frequently needed data in memory, improving query performance.

How It Works:
	1.	When a query needs a page from the database, the database engine first checks the buffer pool.
	2.	If the page is in memory (cache hit), it is retrieved quickly.
	3.	If the page is not in memory (cache miss), it is read from disk and placed in the buffer pool.
	4.	If the buffer pool is full, an existing page is evicted using a replacement policy (e.g., LRU - Least Recently Used).
	5.	Modified pages (dirty pages) are periodically written back to disk (checkpointing or background syncing).

Buffer Pool Advantages:
	•	Minimizes disk I/O by keeping frequently accessed pages in RAM.
	•	Speeds up query execution by reducing the need for slow disk reads.
	•	Manages concurrency efficiently by allowing multiple transactions to work on cached pages.
*/

func (tree *BTree) NewNodePager(fileType string, level uint8) (*NodePager, error) {
	pager, err := NewPager[Node](tree, fileType, level)
	if err != nil {
		return nil, err
	}

	return &NodePager{pager}, nil
}

func (tree *BTree) NewRecordPager(fileType string, level uint8) (*RecordPager, error) {
	pager, err := NewPager[Record](tree, fileType, level)
	if err != nil {
		return nil, err
	}

	return &RecordPager{pager}, nil
}

// Opens or creates a file and sets up the Pager
func NewPager[T Pageable](tree *BTree, fileType string, level uint8) (*Pager[T], error) {
	pageSize := int64(float64(tree.BatchBaseSize) * math.Pow(float64(tree.BatchIncrement)/100, float64(level)))

	headerSize := 0

	path := fmt.Sprintf("SECRETARY/%s/%s_%d_%d.bin", tree.CollectionName, fileType, level, pageSize)
	if fileType == "index" {

		headerSize = SECRETARY_HEADER_LENGTH + int(tree.nodeSize) // Header = SECRETARY_HEADER_LENGTH + ROOT_NODE
		pageSize = 1024 * 1024

		path = fmt.Sprintf("SECRETARY/%s/%s.bin", tree.CollectionName, fileType)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		file.Close()
		// fmt.Printf("Create File : %s\n", path)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}

	{ // Allocate Header
		stat, err := file.Stat()
		if err != nil {
			return nil, err
		}
		if stat.Size() < int64(headerSize) {
			zeroBuf := make([]byte, headerSize)
			_, err = file.WriteAt(zeroBuf, stat.Size())
			if err != nil {
				return nil, err
			}
		}
	}

	pager := &Pager[T]{
		file:       file,
		level:      level,
		headerSize: headerSize,
		pageSize:   pageSize,
		dirtyPages: map[int64]bool{},
	}

	// Initialize Ristretto Cache
	cache, err := ristretto.NewCache(
		&ristretto.Config[int64, *Page[T]]{
			NumCounters: 10000,   // Track frequency of ~10,000 items
			MaxCost:     1 << 20, // 1MB total cache size
			BufferItems: 64,      // Batch writes for performance
			OnEvict: func(item *ristretto.Item[*Page[T]]) {
				delete(pager.dirtyPages, item.Value.Index) // Mark page as clean
			},
		})
	if err != nil {
		return nil, err
	}

	pager.cache = cache

	return pager, nil
}

// AllocatePage writes zeroed data in chunks of pageSize for alignment
func (store *Pager[T]) AllocatePage(index int32) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// Align to the next page boundary
	if ((fileSize - int64(store.headerSize)) % store.pageSize) > 0 {
		return ErrorFileNotAligned(fileInfo)
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.pageSize*int64(index))
	_, err = store.file.WriteAt(zeroBuf, fileSize)
	if err != nil {
		return err
	}

	return nil
}

func (store *Pager[T]) NumPages(index int32) (int64, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return -1, err
	}
	fileSize := fileInfo.Size()

	return (fileSize - int64(store.headerSize)) / store.pageSize, nil
}

/**
* Header
* |
* |
*-----
*
*
*
*-----
*
* Offset
*          +
*-----     + Data shoud not exceed batch boundary
*          +
*
*
*-----
*
**/
// WriteAt writes data at the specified offset in the file.
// If there is not enough free space, it allocates a new batch.
func (store *Pager[T]) WriteAt(data []byte, offset int64) error {
	// Ensure data size does not exceed pageSize
	if ((int64(len(data)) + offset - int64(store.headerSize)) / store.pageSize) !=
		((offset - int64(store.headerSize)) / store.pageSize) {
		return ErrorDataExceedPageSize(len(data), store.pageSize, offset)
	}

	{ // Get current file size
		fileInfo, err := store.file.Stat()
		if err != nil {
			return ErrorFileStat(err)
		}
		fileSize := fileInfo.Size()

		n := 1 + (offset+int64(len(data))-fileSize)/store.pageSize

		// If the requested offset is beyond the current file size, allocate a new batch
		if offset+int64(len(data)) > fileSize {
			err := store.AllocatePage(int32(n))
			if err != nil {
				return ErrorAllocatingBatch(err)
			}
		}
	}

	{
		// Write data at the given offset
		n, err := store.file.WriteAt(data, offset)
		if err != nil || (len(data)) != int(n) {
			return ErrorWritingDataAtOffset(offset, err)
		}
	}

	return nil
}

// ReadAt reads data from the specified offset in the file
func (store *Pager[T]) ReadAt(offset int64, size int32) ([]byte, error) {
	// Allocate a buffer to hold the data
	data := make([]byte, size)

	// Read data from the given offset
	n, err := store.file.ReadAt(data, offset)
	if err != nil || n != int(size) {
		return nil, ErrorReadingDataAtOffset(offset, err)
	}

	return data, nil
}

// func (store *Pager[T]) ReadPage(index int64) (*Page[T], error) {
// 	store.mu.Lock()

// 	// Check if page exists in Ristretto cache
// 	if cachedPage, found := store.cache.Get(index); found {
// 		return cachedPage, nil
// 	}

// 	page := &Page[T]{
// 		Index: index,
// 	}
// 	// Store in Ristretto cache
// 	store.cache.Set(index, page, store.pageSize) // Cost = size of page
// 	store.cache.Wait()                           // Ensure writes are processed

// 	store.mu.Unlock()

// 	page.mu.Lock()
// 	defer page.mu.Unlock()

// 	data, err := store.ReadAt(index*store.pageSize, int32(store.pageSize))
// 	if err != nil {
// 		return nil, err
// 	}

// 	page.Data = data

// 	return page, nil
// }

// // SyncPage writes a page to disk if it's dirty.
// func (store *Pager[T]) SyncPage(index int64) error {
// 	// Get page from cache
// 	page, err := store.ReadPage(index)
// 	if err != nil {
// 		return err
// 	}

// 	page.mu.Lock()
// 	defer page.mu.Unlock()

// 	if _, exist := store.dirtyPages[page.Index]; !exist {
// 		return nil // No need to write unchanged pages
// 	}

// 	offset := index * store.pageSize
// 	err = store.WriteAt(page.Data, offset)
// 	if err != nil {
// 		return err
// 	}

// 	delete(store.dirtyPages, index) // Mark page as clean

// 	return nil
// }

// // Sync writes all dirty pages to disk.
// func (store *Pager[T]) Sync() error {
// 	store.mu.Lock()
// 	defer store.mu.Unlock()

// 	for index := range store.dirtyPages {
// 		if err := store.SyncPage(index); err != nil {
// 			return err
// 		}
// 	}
// 	store.dirtyPages = make(map[int64]bool) // Reset dirty pages

// 	return nil
// }

// // Close syncs pages and closes the file.
// func (store *Pager[T]) Close() error {
// 	if err := store.Sync(); err != nil {
// 		return err
// 	}

// 	store.cache.Close()

// 	return store.file.Close()
// }
