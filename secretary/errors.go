package secretary

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrorInvalidDataLocation = errors.New("Invalid data location")

	ErrorNodeNotInTree              = errors.New("Node not in tree")
	ErrorNodeIsEitherLeaforInternal = errors.New("Node Is Either Leaf or Internal, Node can either have children or record")

	ErrorTreeNotFound = errors.New("Tree not found")
	ErrorTreeNil      = errors.New("Tree nil")

	// Keys
	ErrorKeyNotFound         = errors.New("Key not found")
	ErrorKeyNotInNode        = errors.New("Key not in node")
	ErrorDuplicateKey        = errors.New("Duplicate key")
	ErrorInvalidKey          = errors.New("Invalid key size")
	ErrorKeysNotOrdered      = errors.New("Keys not ordered")
	ErrorKeysGTEOrder        = errors.New("len(n.Keys) >= int(tree.Order)")
	ErrorKeysLTOrder         = errors.New("len(n.Keys) < minKeys")
	ErrorInternalLenChildren = errors.New("len(n.children) != (len(n.Keys) + 1)")
	ErrorNodeMinKeyMismatch  = errors.New("Keys should be minKey of child nodes after first child")

	ErrorLeafLenRecords    = errors.New("len(n.records) != len(n.Keys)")
	ErrorRecordKeyMismatch = errors.New("record.key != key")
	ErrorRecordsNotSorted  = errors.New("Records not sorted")

	ErrorInvalidOrder          = fmt.Errorf("Order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	ErrorInvalidIncrement      = errors.New("Increment must be between 110 and 200")
	ErrorInvalidCollectionName = errors.New("Collection name is not valid, should be a-z 0-9 and with >4 & <30 characters")

	ErrorModeWASM = errors.New("Function disabled : WASM_MODE")

	// File I/O
	ErrorFileNotAligned = func(fileInfo os.FileInfo) error {
		return fmt.Errorf("Error : File %s not aligned", fileInfo.Name())
	}
	ErrorReadingDataAtOffset = func(offset int64, err error) error {
		return fmt.Errorf("Error reading data at offset %d: %v", offset, err)
	}
	ErrorWritingDataAtOffset = func(offset int64, err error) error {
		return fmt.Errorf("Error writing data at offset %d: %v", offset, err)
	}
	ErrorAllocatingBatch = func(err error) error {
		return fmt.Errorf("Error allocating batch: %v", err)
	}
	ErrorFileStat = func(err error) error {
		return fmt.Errorf("Error file stat: %v", err)
	}
	ErrorDataExceedPageSize = func(len int, pageSize int64, offset int64) error {
		return fmt.Errorf("Error: Data size %d exceeds batch size %d at offset %d", len, pageSize, offset)
	}

	// Pointer links
	ErrorParentNotKnowChild = func(child *Node) error {
		return fmt.Errorf("Parent[%d] doesnt know Child[%d]", child.parent.NodeID, child.NodeID)
	}
	ErrorNextNodeLink = func(node *Node) error {
		return fmt.Errorf("Next[%d] doesnt know Node[%d], Next.Prev[%d]", node.next.NodeID, node.NodeID, node.next.prev.NodeID)
	}
	ErrorPrevNodeLink = func(node *Node) error {
		return fmt.Errorf("Prev[%d] doesnt know Node[%d], Prev.Next[%d]", node.prev.NodeID, node.NodeID, node.prev.next.NodeID)
	}
	ErrorChildNotKnowParent = func(parent, child *Node) error {
		return fmt.Errorf("Child[%d] doesnt know Parent[%d], Current Child.Parent[%d]", child.NodeID, parent.NodeID, child.parent.NodeID)
	}
)
