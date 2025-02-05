package secretary

import "errors"

var (
	ErrorNumKeysMoreThanOrder = errors.New("NumKeys cannot be more than order of tree")
	ErrorNumKeysNotMatching   = errors.New("NumKeys not matching")
	ErrorIncorrectKeySize     = errors.New("Incorrect key size")
	ErrorInvalidDataLocation  = errors.New("Invalid data location")
)
