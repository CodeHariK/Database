package secretary

import (
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"sync"
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

// Opens or creates a file and sets up the BatchStore
func (tree *BTree) NewBatchStore(fileType string, level uint8) (*BatchStore, error) {
	batchSize := uint32(float64(tree.BatchBaseSize) * math.Pow(float64(tree.BatchIncrement)/100, float64(level)))

	headerSize := 0

	path := fmt.Sprintf("SECRETARY/%s/%s_%d_%d.bin", tree.CollectionName, fileType, level, batchSize)
	if fileType == "index" {

		headerSize = SECRETARY_HEADER_LENGTH + int(tree.nodeSize) // Header = SECRETARY_HEADER_LENGTH + ROOT_NODE
		batchSize = 1024 * 1024

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

	// fmt.Printf("Open File : %s\n", path)

	return &BatchStore{
		file:       file,
		level:      level,
		headerSize: headerSize,
		batchSize:  batchSize,
	}, nil
}

// AllocateBatch writes zeroed data in chunks of pageSize for alignment
func (store *BatchStore) AllocateBatch(numBatch int32) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return err
	}
	fileSize := fileInfo.Size()

	// Align to the next page boundary
	if (fileSize % int64(store.batchSize)) != 0 {
		//
		//**************
		//
		// -- CHECK : Didnt include headersize
		//
		//**************
		//

		return ErrorFileNotAligned(fileInfo)
	}

	// Expand file by writing zeros
	zeroBuf := make([]byte, store.batchSize*uint32(numBatch))
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
* |
* |
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
func (store *BatchStore) WriteAt(offset int64, data []byte) error {
	// Ensure data size does not exceed batchSize
	if ((int64(len(data)) + offset - int64(store.headerSize)) / int64(store.batchSize)) != ((offset - int64(store.headerSize)) / int64(store.batchSize)) {
		return ErrorDataExceedBatchSize(len(data), store.batchSize, offset)
	}

	// Get current file size
	fileInfo, err := store.file.Stat()
	if err != nil {
		return ErrorFileStat(err)
	}
	fileSize := fileInfo.Size()

	n := 1 + (offset+int64(len(data))-fileSize)/int64(store.batchSize)

	// If the requested offset is beyond the current file size, allocate a new batch
	if offset+int64(len(data)) > fileSize {
		err := store.AllocateBatch(int32(n))
		if err != nil {
			return ErrorAllocatingBatch(err)
		}
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	// Write data at the given offset
	_, err = store.file.WriteAt(data, offset)
	if err != nil {
		return ErrorWritingDataAtOffset(offset, err)
	}

	return nil
}

// ReadAt reads data from the specified offset in the file
func (store *BatchStore) ReadAt(offset int64, size int32) ([]byte, error) {
	store.mu.Lock()
	defer store.mu.Unlock()

	// Allocate a buffer to hold the data
	data := make([]byte, size)
	// data := make([]byte, store.batchSize)

	// Read data from the given offset
	_, err := store.file.ReadAt(data, offset)
	if err != nil {
		return nil, ErrorReadingDataAtOffset(offset, err)
	}

	return data, nil
}

//-------------

type randomAccessFile interface {
	io.ReaderAt
	io.WriterAt
	io.Closer

	Sync() error
	Stat() (fs.FileInfo, error)
	Truncate(size int64) error

	Lock() error
	Unlock() error
}

func readPage(file io.ReaderAt, pageNum uint32, pageSize int) ([]byte, error) {
	offset := int64(pageNum) * int64(pageSize)
	data := make([]byte, pageSize)

	_, err := file.ReadAt(data, offset)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func writePage(file io.WriterAt, pageNum uint32, data []byte, pageSize int) error {
	offset := int64(pageNum) * int64(pageSize)
	_, err := file.WriteAt(data, offset)
	return err
}

//-------------

const (
	PageSize = 4096 // 4KB page size (default in SQLite)
)

// Page represents a single fixed-size page in memory.
type Page struct {
	Number uint32
	Data   []byte
	Dirty  bool // If true, needs to be written back to disk
}

// Pager manages reading and writing pages.
type Pager struct {
	file      *os.File
	pageCache map[uint32]*Page // In-memory cache
	mu        sync.Mutex       // Ensure thread safety
}

// OpenPager opens a database file and initializes the pager.
func OpenPager(filename string) (*Pager, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0o666)
	if err != nil {
		return nil, err
	}
	return &Pager{
		file:      file,
		pageCache: make(map[uint32]*Page),
	}, nil
}

// // ReadPage loads a page from disk into memory.
// func (p *Pager) ReadPage(pageNum uint32) (*Page, error) {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	// Check if page is in cache
// 	if page, exists := p.pageCache[pageNum]; exists {
// 		return page, nil
// 	}

// 	// Allocate a new page buffer
// 	data := make([]byte, PageSize)
// 	offset := int64(pageNum) * PageSize

// 	_, err := p.file.ReadAt(data, offset)
// 	if err != nil && err != os.Eof {
// 		return nil, err
// 	}

// 	// Create a new page object
// 	page := &Page{Number: pageNum, Data: data}
// 	p.pageCache[pageNum] = page
// 	return page, nil
// }

// WritePage writes a page to disk if it's dirty.
func (p *Pager) WritePage(page *Page) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !page.Dirty {
		return nil // No need to write unchanged pages
	}

	offset := int64(page.Number) * PageSize
	_, err := p.file.WriteAt(page.Data, offset)
	if err != nil {
		return err
	}

	page.Dirty = false // Mark as clean after writing
	return nil
}

// AllocatePage creates a new blank page.
func (p *Pager) AllocatePage() (*Page, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Get the next available page number
	stat, err := p.file.Stat()
	if err != nil {
		return nil, err
	}
	pageNum := uint32(stat.Size() / PageSize)

	// Initialize blank page
	data := make([]byte, PageSize)
	page := &Page{Number: pageNum, Data: data, Dirty: true}

	// Add to cache
	p.pageCache[pageNum] = page
	return page, nil
}

// Flush writes all dirty pages to disk.
func (p *Pager) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, page := range p.pageCache {
		if err := p.WritePage(page); err != nil {
			return err
		}
	}
	return nil
}

// Close flushes pages and closes the file.
func (p *Pager) Close() error {
	if err := p.Flush(); err != nil {
		return err
	}
	return p.file.Close()
}
