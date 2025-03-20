package secretary

import "encoding/json"

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

// func (nodes NodeBox) ToBytes() ([]byte, error) {
// 	return binstruct.Serialize(nodes)
// }

// func (records RecordBox) ToBytes() ([]byte, error) {
// 	return binstruct.Serialize(records)
// }

// func (nodes NodeBox) FromBytes(data []byte) error {
// 	err := binstruct.Deserialize(data, nodes)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (records RecordBox) FromBytes(data []byte) error {
// 	err := binstruct.Deserialize(data, records)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (n *NodeBox) ToBytes() ([]byte, error) {
	// Implement serialization logic
	return json.Marshal(n)
}

func (n *NodeBox) FromBytes(data []byte) error {
	// Implement deserialization logic
	var nodes []Node
	err := json.Unmarshal(data, &nodes)
	*n = nodes
	return err
}

func (r *RecordBox) ToBytes() ([]byte, error) {
	// Implement serialization logic
	return json.Marshal(r)
}

func (r *RecordBox) FromBytes(data []byte) error {
	// Implement deserialization logic
	var records []Record
	err := json.Unmarshal(data, &records)
	*r = records
	return err
}
