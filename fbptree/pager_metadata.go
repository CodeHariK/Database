package fbptree

import (
	"fmt"
	"io"
	"io/fs"
)

// the size of the first metadata block in the file,
// reserved for different needs
const (
	metadataSize           = 1000
	customMetadataPosition = 500
)

type randomAccessFile interface {
	io.ReaderAt
	io.WriterAt
	io.Closer

	Sync() error
	Stat() (fs.FileInfo, error)
	Truncate(size int64) error
}

type metadata struct {
	pageSize uint16

	custom []byte
}

func encodeMetadata(m *metadata) []byte {
	data := make([]byte, metadataSize)

	d := encodeUint16(m.pageSize)
	copy(data[0:len(d)], d)

	if len(m.custom) != 0 {
		s := encodeUint16(uint16(len(m.custom)))
		copy(data[customMetadataPosition:customMetadataPosition+len(s)], s)
		copy(data[customMetadataPosition+len(s):], m.custom)
	}

	return data
}

// decodes and returns metadata from the given byte slice.
func decodeMetadata(data []byte) (*metadata, error) {
	// the first block is the page size, encoded as uint16
	pageSize := decodeUint16(data[0:2])

	customMetadataSize := decodeUint16(data[customMetadataPosition : customMetadataPosition+2])
	var customMetadata []byte = nil
	if customMetadataSize != 0 {
		customMetadata = data[customMetadataPosition+2 : customMetadataPosition+2+customMetadataSize]
	}

	return &metadata{pageSize: pageSize, custom: customMetadata}, nil
}

// reads and decodes metadata from the specified file.
func readMetadata(r io.ReaderAt) (*metadata, error) {
	data := make([]byte, metadataSize)
	if read, err := r.ReadAt(data[:], 0); err != nil {
		return nil, fmt.Errorf("failed to read metadata from the file: %w", err)
	} else if read != metadataSize {
		return nil, fmt.Errorf("failed to read metadata from the file: read %d bytes, but must %d", read, metadataSize)
	}

	m, err := decodeMetadata(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	return m, nil
}

func writeMetadata(w io.WriterAt, metadata *metadata) error {
	data := encodeMetadata(metadata)
	if n, err := w.WriteAt(data, 0); err != nil {
		return fmt.Errorf("failed to write the metadata to the file: %w", err)
	} else if n < len(data) {
		return fmt.Errorf("failed to write all the data to the file, wrote %d bytes: %w", n, err)
	}

	return nil
}
