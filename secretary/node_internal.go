package secretary

import "encoding/binary"

// Internal node serialization
func (n *Node) serializeInternal() []byte {
	buf := make([]byte, ChunkSize)
	buf[IsLeafOffset] = 0
	binary.LittleEndian.PutUint32(buf[NumKeysOffset:], uint32(n.NumKeys))
	binary.LittleEndian.PutUint64(buf[ParentIDOffset:], uint64(n.ParentID))

	// Serialize keys
	for i := 0; i < MaxOrder-1; i++ {
		if i < n.NumKeys {
			binary.LittleEndian.PutUint32(buf[InternalKeysStart+i*4:], uint32(n.Keys[i]))
		}
	}

	// Serialize children
	for i := 0; i < MaxOrder; i++ {
		if i <= n.NumKeys {
			binary.LittleEndian.PutUint64(buf[InternalChildStart+i*8:], uint64(n.Children[i]))
		}
	}

	return buf
}
