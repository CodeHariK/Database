package secretary

import (
	"github.com/codeharik/secretary/utils/binstruct"
)

func (datalocation DataLocation) toRecordLocation() RecordLocation {
	return RecordLocation{
		batchLevel: uint8((uint64(datalocation) & RECORD_BATCH_LEVEL_AND) >> 56),
		offset:     uint64(datalocation) & RECORD_BATCH_OFFSET_AND,
	}
}

func (datalocation DataLocation) toNodeLocation() NodeLocation {
	return NodeLocation{
		index:       uint16((uint64(datalocation) & NODE_INDEX_AND) >> 48),
		batchOffset: uint64(datalocation) & NODE_BATCH_OFFSET_AND,
	}
}

func (nodes *Node) ToBytes() ([]byte, error) {
	return binstruct.Serialize(nodes)
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
