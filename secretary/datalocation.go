package secretary

type RecordLocation struct {
	offset     uint64
	batchLevel uint8
}

func (datalocation DataLocation) toRecordLocation() RecordLocation {
	return RecordLocation{
		offset:     uint64(datalocation) & RECORD_BATCH_OFFSET_AND,
		batchLevel: uint8((uint64(datalocation) & RECORD_BATCH_LEVEL_AND) >> 56),
	}
}

type NodeLocation struct {
	batchOffset uint64
	index       uint16
}

func (datalocation DataLocation) toNodeLocation() NodeLocation {
	return NodeLocation{
		batchOffset: uint64(datalocation) & NODE_BATCH_OFFSET_AND,
		index:       uint16((uint64(datalocation) & NODE_INDEX_AND) >> 48),
	}
}

func (tree *BTree) dataLocationCheck(location DataLocation) error {
	if location == -1 {
		return ErrorInvalidDataLocation
	}
	return nil
}
