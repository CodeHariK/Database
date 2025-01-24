package fbptree

import (
	"fmt"
	"io"
)

type freePage struct {
	pageId uint32
	ids    map[uint32]struct{}
	// 0 if does not exist
	nextPageId uint32
}

func (p *freePage) copy() *freePage {
	newIds := make(map[uint32]struct{})
	for key, value := range p.ids {
		newIds[key] = value
	}

	return &freePage{
		p.pageId,
		newIds,
		p.nextPageId,
	}
}

// encodeFreePage encodes free page identifiers into the chunks of byte slices.
func encodeFreePage(page *freePage, pageSize uint16) []byte {
	data := make([]byte, pageSize)
	copy(data[len(data)-pageIdSize:], encodeUint32(page.nextPageId))

	i := 0
	for freePageId := range page.ids {
		copy(data[i:], encodeUint32(freePageId))
		i += pageIdSize
	}

	return data
}

func decodeFreePage(pageId uint32, data []byte) (*freePage, error) {
	pageIdNum := (len(data) - pageIdSize) / pageIdSize
	freePages := make(map[uint32]struct{})
	for i := 0; i < pageIdNum; i++ {
		from, to := i*pageIdSize, i*pageIdSize+pageIdSize

		pageId := decodeUint32(data[from:to])
		if pageId == 0 {
			break
		}

		freePages[pageId] = struct{}{}
	}

	nextPageId := decodeUint32(data[len(data)-pageIdSize:])

	return &freePage{pageId, freePages, nextPageId}, nil
}

func readFreePage(r io.ReaderAt, pageId uint32, pageSize uint16) (*freePage, error) {
	data, err := readPage(r, pageId, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to read page %d: %w", pageId, err)
	}

	freePage, err := decodeFreePage(pageId, data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode free page: %w", err)
	}

	return freePage, nil
}

// readFreePages reads and initializes the list of free pages.
func readFreePages(r io.ReaderAt, pageSize uint16) (map[uint32]*freePage, *freePage, map[uint32]*freePage, map[uint32]uint32, error) {
	isFreePage := make(map[uint32]*freePage)
	freePages := make(map[uint32]*freePage)
	prevPageIds := make(map[uint32]uint32)

	var prevPageId uint32
	freePageId := firstFreePageId
	var lastFreePage *freePage
	for freePageId != 0 {
		freePage, err := readFreePage(r, freePageId, pageSize)
		if err != nil {
			return nil, nil, nil, nil, fmt.Errorf("failed to read free page: %w", err)
		}

		for id := range freePage.ids {
			isFreePage[id] = freePage
		}
		freePages[freePageId] = freePage

		if prevPageId != 0 {
			prevPageIds[freePageId] = prevPageId
		}
		prevPageId = freePageId

		lastFreePage = freePage
		freePageId = freePage.nextPageId
	}

	return isFreePage, lastFreePage, freePages, prevPageIds, nil
}

func writePage(w io.WriterAt, pageId uint32, data []byte, pageSize uint16) error {
	offset := int64(metadataSize + (pageId-1)*uint32(pageSize))

	if n, err := w.WriteAt(data, offset); err != nil {
		return fmt.Errorf("failed to write the page: %w", err)
	} else if n != len(data) {
		return fmt.Errorf("failed to write %d bytes, wrote %d", len(data), n)
	}

	return nil
}

func readPage(r io.ReaderAt, pageId uint32, pageSize uint16) ([]byte, error) {
	offset := int64(metadataSize + (pageId-1)*uint32(pageSize))
	data := make([]byte, pageSize)
	if n, err := r.ReadAt(data, offset); err != nil {
		return nil, fmt.Errorf("failed to read the page data: %w", err)
	} else if n != int(pageSize) {
		return nil, fmt.Errorf("failed to read %d bytes, read %d", pageSize, n)
	}

	return data, nil
}
