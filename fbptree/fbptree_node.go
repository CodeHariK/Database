package fbptree

import "fmt"

// node reprents a node in the B+ tree.
type node struct {
	id uint32

	// true for leaf node and root without children
	// and false for internal node and root with children
	leaf     bool
	parentID uint32

	// Real key number is stored under the keyNum.
	keys   [][]byte
	keyNum int

	// Leaf nodes can point to the value,
	// but internal nodes point to the nodes. So
	// to save space, we can use pointers abstraction.
	// The size of pointers equals to the size of keys + 1.
	// In the leaf node, the last pointers element points to
	// the next leaf node.
	pointers []*pointer
}

func encodeNode(node *node) []byte {
	data := make([]byte, 0)

	data = append(data, encodeUint32(node.id)...)
	data = append(data, encodeUint32(node.parentID)...)
	data = append(data, encodeBool(node.leaf)...)
	data = append(data, encodeUint16(uint16(node.keyNum))...)
	data = append(data, encodeUint16(uint16(len(node.keys)))...)

	for _, key := range node.keys {
		if key == nil {
			break
		}

		data = append(data, encodeUint16(uint16(len(key)))...)
		data = append(data, key...)
	}

	pointerNum := node.keyNum
	if !node.leaf {
		pointerNum += 1
	}

	data = append(data, encodeUint16(uint16(pointerNum))...)
	data = append(data, encodeUint16(uint16(len(node.pointers)))...)
	for i := 0; i < pointerNum; i++ {
		pointer := node.pointers[i]
		if pointer.isNodeID() {
			data = append(data, 0)
			data = append(data, encodeUint32(pointer.asNodeID())...)
		} else if pointer.isValue() {
			data = append(data, 1)
			data = append(data, encodeUint16(uint16(len(pointer.asValue())))...)
			data = append(data, pointer.asValue()...)
		}
	}

	var nextID uint32
	if node.next() != nil {
		nextID = node.next().asNodeID()
		data = append(data, encodeBool(true)...)
		data = append(data, encodeUint32(nextID)...)
	} else {
		data = append(data, encodeBool(false)...)
		data = append(data, 0)
	}

	return data
}

func decodeNode(data []byte) (*node, error) {
	position := 0
	nodeID := decodeUint32(data[position : position+4])
	position += 4
	parentID := decodeUint32(data[position : position+4])
	position += 4
	leaf := decodeBool(data[position : position+1])
	position += 1

	keyNum := decodeUint16(data[position : position+2])
	position += 2
	keyLen := int(decodeUint16(data[position : position+2]))
	position += 2
	keys := make([][]byte, keyLen)
	for k := 0; k < int(keyNum); k++ {
		keySize := int(decodeUint16(data[position : position+2]))
		position += 2

		key := data[position : position+keySize]
		keys[k] = key
		position += keySize
	}

	pointerNum := decodeUint16(data[position : position+2])
	position += 2
	pointerLen := int(decodeUint16(data[position : position+2]))
	position += 2
	pointers := make([]*pointer, pointerLen)
	for p := 0; p < int(pointerNum); p++ {
		if data[position] == 0 {
			position += 1
			// nodeID

			nodeID := decodeUint32(data[position : position+4])
			position += 4

			pointers[p] = &pointer{nodeID}
		} else if data[position] == 1 {
			position += 1
			// value

			valueSize := int(decodeUint16(data[position : position+2]))
			position += 2

			value := data[position : position+valueSize]
			position += valueSize

			pointers[p] = &pointer{value}
		}
	}

	n := &node{
		nodeID,
		leaf,
		parentID,
		keys,
		int(keyNum),
		pointers,
	}

	hasNextID := decodeBool(data[position : position+1])
	position += 1

	if hasNextID {
		nextID := decodeUint32(data[position : position+4])
		n.setNext(&pointer{nextID})
	}

	return n, nil
}

// append apppends key and the pointer to the node
func (n *node) append(key []byte, p *pointer, storage *storage) error {
	keyPosition := n.keyNum
	pointerPosition := n.keyNum
	if !n.leaf && n.pointers[pointerPosition] != nil {
		pointerPosition++
	}

	n.keys[keyPosition] = key
	n.pointers[pointerPosition] = p
	n.keyNum++

	if !n.leaf {
		childID := p.asNodeID()
		child, err := storage.loadNodeByID(childID)
		if err != nil {
			return fmt.Errorf("failed load the child node %d: %w", childID, err)
		}

		child.parentID = n.id

		err = storage.updateNodeByID(childID, child)
		if err != nil {
			return fmt.Errorf("failed to update the child node %d: %w", childID, err)
		}
	}

	return nil
}

// copyFromRight copies the keys and the pointer from the given node.
func (n *node) copyFromRight(from *node, storage *storage) error {
	for i := 0; i < from.keyNum; i++ {
		err := n.append(from.keys[i], from.pointers[i], storage)
		if err != nil {
			return fmt.Errorf("failed to append to %d: %w", n.id, err)
		}
	}

	if n.leaf {
		n.setNext(from.next())

		err := storage.updateNodeByID(n.id, n)
		if err != nil {
			return fmt.Errorf("failed to update the node %d: %w", n.id, err)
		}
	} else {
		n.pointers[n.keyNum] = from.pointers[from.keyNum]

		childID := n.pointers[n.keyNum].asNodeID()
		child, err := storage.loadNodeByID(childID)
		if err != nil {
			return fmt.Errorf("failed to load the child node %d: %w", childID, err)
		}

		child.parentID = n.id

		err = storage.updateNodeByID(child.id, child)
		if err != nil {
			return fmt.Errorf("failed to update the parent for the child node %d: %w", childID, err)
		}
	}

	return nil
}

// pointerPositionOf finds the pointer position of the given node.
// Returns -1 if it is not found.
func (n *node) pointerPositionOf(x *node) int {
	for position, pointer := range n.pointers {
		if pointer == nil {
			// reached the end
			break
		}

		if pointer.asNodeID() == x.id {
			return position
		}
	}

	// pointer not found
	return -1
}

// deleteAt deletes the entry at the position and shifts
// the keys and the pointers.
func (n *node) deleteAt(keyPosition int, pointerPosition int) {
	// shift the keys
	for j := keyPosition; j < n.keyNum-1; j++ {
		n.keys[j] = n.keys[j+1]
	}
	n.keys[n.keyNum-1] = nil

	pointerNum := n.keyNum
	if !n.leaf {
		pointerNum++
	}
	// shift the pointers
	for j := pointerPosition; j < pointerNum-1; j++ {
		n.pointers[j] = n.pointers[j+1]
	}
	n.pointers[pointerNum-1] = nil

	n.keyNum--
}

// setNext sets the "next" pointer (the last pointer) to the next node. Only relevant
// for the leaf nodes.
func (n *node) setNext(p *pointer) {
	n.pointers[len(n.pointers)-1] = p
}

// next returns the pointer to the next leaf node. Only relevant
// for the leaf nodes.
func (n *node) next() *pointer {
	return n.pointers[len(n.pointers)-1]
}

// keyPosition returns the position of the key, but -1 if it is not present.
func (n *node) keyPosition(key []byte) int {
	keyPosition := 0
	for ; keyPosition < n.keyNum; keyPosition++ {
		if compare(key, n.keys[keyPosition]) == 0 {
			return keyPosition
		}
	}

	return -1
}

// insertAt inserts the specified key and pointer at the specified position.
// Only works with leaf nodes.
func (n *node) insertAt(keyPosition int, key []byte, pointerPosition int, pointer *pointer) {
	for j := n.keyNum; j > keyPosition; j-- {
		n.keys[j] = n.keys[j-1]
	}

	pointerNum := n.keyNum
	if !n.leaf {
		pointerNum += 1
	}

	for j := pointerNum; j > pointerPosition; j-- {
		n.pointers[j] = n.pointers[j-1]
	}

	n.keys[keyPosition] = key
	n.pointers[pointerPosition] = pointer
	n.keyNum++
}
