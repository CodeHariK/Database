package secretary

import (
	"bytes"
	"testing"
)

func TestPagerAllocateBatch(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	fileInfo, err := tree.nodePager.file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	originalFileSize := fileInfo.Size()

	{ // Allocate the first batch
		err := tree.nodePager.AllocatePage(1)
		if err != nil {
			t.Fatalf("AllocateBatch failed: %v", err)
		}

		fileInfo, err := tree.nodePager.file.Stat()
		allocatedSize := fileInfo.Size() - originalFileSize
		if err != nil || allocatedSize != int64(tree.nodePager.itemSize) {
			t.Fatalf("Expected file size %d, got %d", tree.nodePager.itemSize, fileInfo.Size())
		}
	}

	{ // Allocate another batch
		err := tree.nodePager.AllocatePage(1)
		if err != nil {
			t.Fatalf("AllocateBatch failed on second batch: %v", err)
		}

		// Ensure the file size has increased correctly
		fileInfo, err = tree.nodePager.file.Stat()
		allocatedSize := fileInfo.Size() - originalFileSize
		expectedSize := int64(2 * tree.nodePager.itemSize)
		if err != nil || allocatedSize != expectedSize {
			t.Fatalf("Expected file size %d, got %d", expectedSize, fileInfo.Size())
		}
	}

	s.PagerShutdown()
}

func TestPagerWriteAndReadAtOffset(t *testing.T) {
	s := dummySecretary(t)
	tree := dummyTree(t, s, 10)

	{ // Test: Write small data within the first batch
		data := []byte("Hello, B+ Tree!")
		offset := int64(0)

		err := tree.nodePager.WriteAt(data, offset)
		if err != nil {
			t.Fatalf("WriteAt failed: %v", err)
		}

		fileInfo, err := tree.nodePager.file.Stat()
		if err != nil || fileInfo.Size() != tree.nodePager.headerSize {
			t.Fatalf("Expected file size %d, got %d", tree.nodePager.headerSize, fileInfo.Size())
		}

		// Read the data back
		readData, err := tree.nodePager.ReadAt(offset, int32(len(data)))
		if err != nil {
			t.Fatalf("ReadAtOffset failed: %v", err)
		}

		// Compare written and read data
		if !bytes.Equal(data, readData) {
			t.Fatalf("Expected '%s', got '%s'", data, readData[:len(data)])
		}
	}

	{ // Test: Writing beyond the current file size should allocate more batch
		offset := int64(float64(tree.nodePager.itemSize) * 3.5) // Beyond the first batch (1024)
		data := []byte("Second Batch Data")

		err := tree.nodePager.WriteAt(data, offset)
		if err != nil {
			t.Fatalf("WriteAt failed on batch allocation: %v", err)
		}

		// Read the new data back
		readData, err := tree.nodePager.ReadAt(offset, int32(len(data)))
		if err != nil {
			t.Fatalf("ReadAtOffset failed: %v", err)
		}

		// Compare written and read data for the second batch
		if !bytes.Equal(data, readData) {
			t.Fatalf("Expected '%s' at offset %d, got '%s'", data, offset, readData[:len(data)])
		}

		fileInfo, err := tree.nodePager.file.Stat()
		if err != nil || fileInfo.Size() != tree.nodePager.headerSize+4*int64(tree.nodePager.itemSize) {
			t.Fatalf("Expected file size %d, got %d", 4*int64(tree.nodePager.itemSize), fileInfo.Size())
		}
	}

	{ // Test: Write data that exceed batch size, Should Fail
		offset := int64(float64(tree.nodePager.itemSize) * 5.8)
		data := make([]byte, int64(float64(tree.nodePager.itemSize)*0.8))

		err := tree.nodePager.WriteAt(data, offset)
		if err == nil {
			t.Fatalf("WriteAt should be failing, data should not exceed batch : %v", err)
		}
	}

	s.PagerShutdown()
}
