package secretary

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrorNumKeysMoreThanOrder       = errors.New("NumKeys cannot be more than order of tree")
	ErrorNumKeysNotMatching         = errors.New("NumKeys not matching")
	ErrorInvalidDataLocation        = errors.New("Invalid data location")
	ErrorNodeNotInTree              = errors.New("Node not in tree")
	ErrorNodeIsEitherLeaforInternal = errors.New("Node Is Either Leaf or Internal, Node can either have children or record")

	ErrorKeyNotFound  = errors.New("Key not found")
	ErrorKeyNotInNode = errors.New("Key not in node")
	ErrorDuplicateKey = errors.New("Duplicate key")
	ErrorInvalidKey   = errors.New("Invalid key size")

	ErrorInvalidOrder          = fmt.Errorf("Order must be between %d and %d", MIN_ORDER, MAX_ORDER)
	ErrorInvalidBatchIncrement = fmt.Errorf("Batch Increment must be between 110 and 200")

	ErrorInvalidCollectionName = func() error {
		return fmt.Errorf("Collection name is not valid, should be a-z 0-9 and with >4 & <30 characters")
	}

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

	ErrorDataExceedBatchSize = func(len int, batchSize uint32, offset int64) error {
		return fmt.Errorf("Error: Data size %d exceeds batch size %d at offset %d", len, batchSize, offset)
	}
)
