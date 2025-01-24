package fbptree

import (
	"encoding/binary"
)

func decodeUint16(data []byte) uint16 {
	return binary.BigEndian.Uint16(data)
}

func encodeUint16(v uint16) []byte {
	var data [2]byte
	binary.BigEndian.PutUint16(data[:], v)

	return data[:]
}

func decodeUint32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

func encodeUint32(v uint32) []byte {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], v)

	return data[:]
}

func encodeBool(v bool) []byte {
	var data [1]byte
	if v {
		data[0] = 1
	}

	return data[:]
}

func decodeBool(data []byte) bool {
	return data[0] == 1
}

//-------------------------------------------------------------------

type config struct {
	order    uint16
	pageSize uint16
}

//-------------------------------------------------------------------

type treeMetadata struct {
	order      uint16
	rootID     uint32
	leftmostID uint32
	size       uint32
}

func encodeTreeMetadata(metadata *treeMetadata) []byte {
	var data [14]byte

	copy(data[0:2], encodeUint16(metadata.order))
	copy(data[2:6], encodeUint32(metadata.rootID))
	copy(data[6:10], encodeUint32(metadata.leftmostID))
	copy(data[10:14], encodeUint32(metadata.size))

	return data[:]
}

func decodeTreeMetadata(data []byte) (*treeMetadata, error) {
	return &treeMetadata{
		order:      decodeUint16(data[0:2]),
		rootID:     decodeUint32(data[2:6]),
		leftmostID: decodeUint32(data[6:10]),
		size:       decodeUint32(data[10:14]),
	}, nil
}

//-------------------------------------------------------------------

// pointer wraps the node or the value.
type pointer struct {
	value interface{}
}

func (p *pointer) isNodeID() bool {
	_, ok := p.value.(uint32)

	return ok
}

func (p *pointer) isValue() bool {
	_, ok := p.value.([]byte)

	return ok
}

// asNode returns a node ID.
func (p *pointer) asNodeID() uint32 {
	return p.value.(uint32)
}

// asValue returns a asValue instance of the value.
func (p *pointer) asValue() []byte {
	return p.value.([]byte)
}

// overrideValue overrides the value
func (p *pointer) overrideValue(newValue []byte) []byte {
	oldValue := p.value.([]byte)
	p.value = newValue

	return oldValue
}
