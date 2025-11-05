package file

import (
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
)

const (
	ChunkSize = 1024 // Adjust chunk size
)

// Metadata structure
type Metadata struct {
	Filename  string   `json:"filename"`
	FileSize  int64    `json:"file_size"`
	NumChunks int32    `json:"num_chunks"`
	Chunks    []string `json:"chunks"` // "chunkname:index_hash"
}

func splitFile(filePath string, metadataFile string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()

	metadir := filepath.Dir(metadataFile)

	// Ensure chunk directory exists
	os.MkdirAll(metadir, os.ModePerm)

	metadata := Metadata{
		Filename:  filePath,
		FileSize:  fileSize,
		NumChunks: int32(fileSize/ChunkSize) + 1,
		Chunks:    make([]string, int32(fileSize/ChunkSize)+1),
	}

	buffer := make([]byte, ChunkSize)
	index := 0

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hash := crc32.ChecksumIEEE(buffer[:n])

			// Store metadata with format: "chunkname:index_hash"
			metadata.Chunks[index] = fmt.Sprintf("%d_%08x", index, hash)

			// Save chunk
			if err := os.WriteFile(filepath.Join(metadir, metadata.Chunks[index]), buffer[:n], 0o644); err != nil {
				return err
			}

			index++
		}
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}

	// Save metadata
	metaDataBytes, _ := json.MarshalIndent(metadata, "", "  ")
	return os.WriteFile(metadataFile, metaDataBytes, 0o644)
}

func mergeChunks(metadataFile string, reconstructedFile string) error {
	// Load metadata
	metaDataBytes, err := os.ReadFile(metadataFile)
	if err != nil {
		return err
	}

	var metadata Metadata
	if err := json.Unmarshal(metaDataBytes, &metadata); err != nil {
		return err
	}

	// Create output file
	outFile, err := os.Create(reconstructedFile)
	if err != nil {
		return err
	}
	defer outFile.Close()

	metadir := filepath.Dir(metadataFile)

	// Reassemble file from chunks
	for _, chunkName := range metadata.Chunks {
		data, err := os.ReadFile(filepath.Join(metadir, chunkName))
		if err != nil {
			return err
		}

		// Verify integrity
		hash := crc32.ChecksumIEEE(data)
		expectedHash := chunkName[len(chunkName)-8:] // Extract last 8 chars (hash)
		if fmt.Sprintf("%08x", hash) != expectedHash {
			return fmt.Errorf("hash mismatch for chunk: %s", expectedHash)
		}

		// Append to final file
		_, err = outFile.Write(data)
		if err != nil {
			return err
		}
	}

	// Verify final file size
	finalStat, err := os.Stat(reconstructedFile)
	if err != nil {
		return err
	}
	if finalStat.Size() != metadata.FileSize {
		return fmt.Errorf("file size mismatch! Expected: %d, Got: %d", metadata.FileSize, finalStat.Size())
	}

	// // Delete chunks after successful merge
	// for _, chunkName := range metadata.Chunks {
	// 	os.Remove(filepath.Join(metadir, chunkName))
	// }
	// os.Remove(metadataFile)

	return nil
}
