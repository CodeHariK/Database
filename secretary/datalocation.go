package secretary

import "github.com/codeharik/secretary/utils/binstruct"

/*
convert byte[] to records and node
traverse entire tree and store it in disk
Store entire node in same page
Put nodeId,nodeoffset in pagemetadata for records
Store continous node together
Split page when node is added more than nodecapacity of page
Put Page on different batchlevel when exceeding pagesize
*/

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

func (node Node) ToBytes() ([]byte, error) {
	return binstruct.Serialize(node)
}

func (record Record) ToBytes() ([]byte, error) {
	return binstruct.Serialize(record)
}
