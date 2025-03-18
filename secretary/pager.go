package secretary

import (
	"fmt"
	"math"
	"os"
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

*/

// Opens or creates a file and sets up the Pager
func (tree *BTree) NewPager(fileType string, level uint8) (*Pager, error) {
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

	return &Pager{
		file:       file,
		level:      level,
		headerSize: headerSize,
		pageSize:   pageSize,
		pageCache:  make(map[int64]*Page),
	}, nil
}

// AllocatePage writes zeroed data in chunks of pageSize for alignment
func (store *Pager) AllocatePage(numBatch int32) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// Align to the next page boundary
	if ((fileSize - int64(store.headerSize)) % int64(store.pageSize)) > 0 {
		return ErrorFileNotAligned(fileInfo)
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.pageSize*int64(numBatch))
	_, err = store.file.WriteAt(zeroBuf, fileSize)
	if err != nil {
		return err
	}

	return nil
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
func (store *Pager) WriteAt(offset int64, data []byte) error {
	// Ensure data size does not exceed pageSize
	if ((int64(len(data)) + offset - int64(store.headerSize)) / int64(store.pageSize)) !=
		((offset - int64(store.headerSize)) / int64(store.pageSize)) {
		return ErrorDataExceedPageSize(len(data), store.pageSize, offset)
	}

	{ // Get current file size
		fileInfo, err := store.file.Stat()
		if err != nil {
			return ErrorFileStat(err)
		}
		fileSize := fileInfo.Size()

		n := 1 + (offset+int64(len(data))-fileSize)/int64(store.pageSize)

		// If the requested offset is beyond the current file size, allocate a new batch
		if offset+int64(len(data)) > fileSize {
			err := store.AllocatePage(int32(n))
			if err != nil {
				return ErrorAllocatingBatch(err)
			}
		}
	}

	{
		store.mu.Lock()
		defer store.mu.Unlock()

		// Write data at the given offset
		n, err := store.file.WriteAt(data, offset)
		if err != nil || (len(data)) != int(n) {
			return ErrorWritingDataAtOffset(offset, err)
		}
	}

	return nil
}

// ReadAt reads data from the specified offset in the file
func (store *Pager) ReadAt(offset int64, size int32) ([]byte, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Allocate a buffer to hold the data
	data := make([]byte, size)

	// Read data from the given offset
	n, err := store.file.ReadAt(data, offset)
	if err != nil || n != int(size) {
		return nil, ErrorReadingDataAtOffset(offset, err)
	}

	return data, nil
}

func (store *Pager) ReadPage(index int64) (*Page, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Check if page is in cache
	if page, exists := store.pageCache[index]; exists {
		return page, nil
	}

	data, err := store.ReadAt(index*store.pageSize, int32(store.pageSize))
	if err != nil {
		return nil, err
	}

	page := &Page{
		Index: index,
		Data:  data,
		Dirty: false,
	}
	store.pageCache[index] = page

	return page, nil
}

// WritePage writes a page to disk if it's dirty.
func (store *Pager) WritePage(index int64) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	page := store.pageCache[index]

	if !page.Dirty {
		return nil // No need to write unchanged pages
	}

	offset := index * store.pageSize
	_, err := store.file.WriteAt(page.Data, offset)
	if err != nil {
		return err
	}

	page.Dirty = false // Mark as clean after writing
	return nil
}

// Flush writes all dirty pages to disk.
func (store *Pager) Flush() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	for _, page := range store.pageCache {
		if err := store.WritePage(page.Index); err != nil {
			return err
		}
	}
	return nil
}

// Close flushes pages and closes the file.
func (store *Pager) Close() error {
	if err := store.Flush(); err != nil {
		return err
	}
	return store.file.Close()
}
