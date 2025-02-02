package secretary

import "testing"

func TestAllocateBatch(t *testing.T) {
	tree := testBTree(t, "TestAllocateBatch")

	// Allocate the first batch
	err := tree.nodeBatchStore.AllocateBatch(1)
	if err != nil {
		t.Fatalf("AllocateBatch failed: %v", err)
	}

	fileInfo, err := tree.nodeBatchStore.file.Stat()
	if err != nil || fileInfo.Size() != int64(tree.nodeBatchStore.batchSize) {
		t.Errorf("Expected file size %d, got %d", tree.nodeBatchStore.batchSize, fileInfo.Size())
	}

	// Allocate another batch
	err = tree.nodeBatchStore.AllocateBatch(1)
	if err != nil {
		t.Fatalf("AllocateBatch failed on second batch: %v", err)
	}

	// Ensure the file size has increased correctly
	fileInfo, _ = tree.nodeBatchStore.file.Stat()
	expectedSize := int64(2 * tree.nodeBatchStore.batchSize)
	if fileInfo.Size() != expectedSize {
		t.Errorf("Expected file size %d, got %d", expectedSize, fileInfo.Size())
	}
}

func TestWriteAndReadAtOffset(t *testing.T) {
	tree := testBTree(t, "TestWriteAndReadAtOffset")

	fileInfo, _ := tree.nodeBatchStore.file.Stat()
	if fileInfo.Size() != 0 {
		t.Errorf("Expected file size %d, got %d", 0, fileInfo.Size())
	}

	// Test: Write small data within the first batch
	data := []byte("Hello, B+ Tree!")
	offset := int64(0)

	err := tree.nodeBatchStore.WriteAtOffset(offset, data)
	if err != nil {
		t.Fatalf("WriteAtOffset failed: %v", err)
	}

	fileInfo, _ = tree.nodeBatchStore.file.Stat()
	if fileInfo.Size() != int64(tree.nodeBatchStore.batchSize) {
		t.Errorf("Expected file size %d, got %d", tree.nodeBatchStore.batchSize, fileInfo.Size())
	}

	// Read the data back
	readData, err := tree.nodeBatchStore.ReadAtOffset(offset, int32(len(data)))
	if err != nil {
		t.Fatalf("ReadAtOffset failed: %v", err)
	}

	// Compare written and read data
	if string(readData[:len(data)]) != string(data) {
		t.Errorf("Expected '%s', got '%s'", data, readData[:len(data)])
	}

	// Test: Writing beyond the current file size should allocate a new batch
	newOffset := int64(float64(tree.nodeBatchStore.batchSize) * 3.5) // Beyond the first batch (1024)
	data2 := []byte("Second Batch Data")

	err = tree.nodeBatchStore.WriteAtOffset(newOffset, data2)
	if err != nil {
		t.Fatalf("WriteAtOffset failed on batch allocation: %v", err)
	}

	// Read the new data back
	readData2, err := tree.nodeBatchStore.ReadAtOffset(newOffset, int32(len(data2)))
	if err != nil {
		t.Fatalf("ReadAtOffset failed: %v", err)
	}

	// Compare written and read data for the second batch
	if string(readData2[:len(data2)]) != string(data2) {
		t.Errorf("Expected '%s' at offset %d, got '%s'", data2, newOffset, readData2[:len(data2)])
	}

	fileInfo, _ = tree.nodeBatchStore.file.Stat()
	if fileInfo.Size() != 4*int64(tree.nodeBatchStore.batchSize) {
		t.Errorf("Expected file size %d, got %d", 4*int64(tree.nodeBatchStore.batchSize), fileInfo.Size())
	}

	newOffset = int64(float64(tree.nodeBatchStore.batchSize) * 5.8) // Beyond the first batch (1024)
	data3 := make([]byte, int64(float64(tree.nodeBatchStore.batchSize)*0.4))

	err = tree.nodeBatchStore.WriteAtOffset(newOffset, data3)
	if err == nil {
		t.Fatalf("WriteAtOffset should be failing, data should be in same batch: %v", err)
	}
}
