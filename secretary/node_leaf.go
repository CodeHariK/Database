package secretary

// // Leaf node serialization
// func (n *Node) serializeLeaf() []byte {
// 	buf := make([]byte, ChunkSize)
// 	buf[IsLeafOffset] = 1
// 	binary.LittleEndian.PutUint32(buf[NumKeysOffset:], uint32(n.NumKeys))
// 	binary.LittleEndian.PutUint64(buf[ParentIDOffset:], uint64(n.ParentID))
// 	binary.LittleEndian.PutUint64(buf[LeafNextIDOffset:], uint64(n.NextID))

// 	// Serialize keys
// 	for i := 0; i < MaxOrder-1; i++ {
// 		if i < n.NumKeys {
// 			binary.LittleEndian.PutUint32(buf[LeafKeysStart+i*4:], uint32(n.Keys[i]))
// 		}
// 	}

// 	// Serialize record IDs
// 	for i := 0; i < MaxOrder-1; i++ {
// 		if i < n.NumKeys {
// 			binary.LittleEndian.PutUint64(buf[LeafRecordsStart+i*8:], uint64(n.RecordIDs[i]))
// 		}
// 	}

// 	return buf
// }
