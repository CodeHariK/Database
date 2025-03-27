package secretary

import (
	"github.com/codeharik/secretary/utils/binstruct"
)

func ToRecordLocation(datalocation uint64) RecordLocation {
	return RecordLocation{
		batchLevel: uint8((uint64(datalocation) & RECORD_BATCH_LEVEL_AND) >> 56),
		offset:     uint64(datalocation) & RECORD_BATCH_OFFSET_AND,
	}
}

func (nodes *Node) ToBytes() ([]byte, error) {
	return binstruct.Serialize(nodes)
}

func (nodes *Node) NewPage(index int64) *Page[*Node] {
	return &Page[*Node]{
		Index: index,
		Data:  &Node{},
	}
}

func (records *Record) NewPage(index int64) *Page[*Record] {
	return &Page[*Record]{
		Index: index,
		Data:  &Record{},
	}
}

func (nodes *Node) FromBytes(data []byte) error {
	err := binstruct.Deserialize(data, nodes)
	if err != nil {
		return err
	}
	return nil
}

func (records *Record) ToBytes() ([]byte, error) {
	return binstruct.Serialize(records)
}

func (records *Record) FromBytes(data []byte) error {
	err := binstruct.Deserialize(data, records)
	if err != nil {
		return err
	}
	return nil
}
