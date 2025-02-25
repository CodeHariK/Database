package binstruct

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"reflect"
	"sort"
)

// Serialize struct to binary []byte (Little-Endian)
// bin : type name
// byte : number of bytes used for length of string or []byte
// max : max length of string or []byte
// array_elem_len : max length (array elements) in (array of (array elements)), [][]byte [][]int32 [][]float64
func Serialize(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		return nil, errors.New("SerializeBinary: expected a value not pointer")
	}

	buf := new(bytes.Buffer)
	typ := reflect.TypeOf(s)

	sortedFields := getSortedFields(typ)
	for _, field := range sortedFields {

		tag := field.Tag.Get("bin")

		// Skip fields without bin tag
		if tag == "" {
			continue
		}
		fieldValue, numBytes, maxStorableSize, maxSize, array_elem_len := extractFieldParameters(val, field)

		// Handle fields based on their types
		switch fieldValue.Kind() {
		case reflect.Int8:
			binary.Write(buf, binary.LittleEndian, int8(fieldValue.Int()))
		case reflect.Uint8:
			binary.Write(buf, binary.LittleEndian, uint8(fieldValue.Uint()))
		case reflect.Int16:
			binary.Write(buf, binary.LittleEndian, int16(fieldValue.Int()))
		case reflect.Uint16:
			binary.Write(buf, binary.LittleEndian, uint16(fieldValue.Uint()))
		case reflect.Int32:
			binary.Write(buf, binary.LittleEndian, int32(fieldValue.Int()))
		case reflect.Uint32:
			binary.Write(buf, binary.LittleEndian, uint32(fieldValue.Uint()))
		case reflect.Int64:
			binary.Write(buf, binary.LittleEndian, int64(fieldValue.Int()))
		case reflect.Uint64:
			binary.Write(buf, binary.LittleEndian, uint64(fieldValue.Uint()))
		case reflect.Float64:
			binary.Write(buf, binary.LittleEndian, float64(fieldValue.Float()))
		case reflect.String:
			str := fieldValue.String()
			if len(str) >= maxSize {
				str = str[:maxSize]
			}
			if len(str) >= maxStorableSize {
				str = str[:maxStorableSize]
			}
			writeByteLen(buf, numBytes, len(str))
			buf.WriteString(str) // Write string directly
		case reflect.Slice:
			elemKind := fieldValue.Type().Elem().Kind()
			elemType := fieldValue.Type().Elem()

			length := fieldValue.Len()

			// Apply truncation logic
			if length > maxSize {
				length = maxSize
			}
			if length > maxStorableSize {
				length = maxStorableSize
			}

			// Write the truncated length prefix
			writeByteLen(buf, numBytes, length)

			if elemKind == reflect.Uint8 { // Special case for []byte
				data := fieldValue.Bytes()
				buf.Write(data[:length]) // Write directly
			} else if elemKind == reflect.Slice {
				elemBaseKind := elemType.Elem().Kind()
				elemLen := reflectKindByteLen(elemBaseKind)

				for i := 0; i < length; i++ {
					var itemBytes []byte

					switch elemBaseKind {
					case reflect.Uint8, reflect.Int8:
						itemBytes = fieldValue.Index(i).Interface().([]byte)

					case reflect.Int16:
						item := fieldValue.Index(i).Interface().([]int16)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							binary.LittleEndian.PutUint16(itemBytes[j*elemLen:], uint16(v))
						}
					case reflect.Uint16:
						item := fieldValue.Index(i).Interface().([]uint16)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							binary.LittleEndian.PutUint16(itemBytes[j*elemLen:], v)
						}

					case reflect.Int32:
						item := fieldValue.Index(i).Interface().([]int32)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							binary.LittleEndian.PutUint32(itemBytes[j*elemLen:], uint32(v))
						}
					case reflect.Uint32:
						item := fieldValue.Index(i).Interface().([]uint32)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							binary.LittleEndian.PutUint32(itemBytes[j*elemLen:], v)
						}

					case reflect.Int64:
						item := fieldValue.Index(i).Interface().([]int64)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							binary.LittleEndian.PutUint64(itemBytes[j*elemLen:], uint64(v))
						}
					case reflect.Uint64:
						item := fieldValue.Index(i).Interface().([]uint64)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							binary.LittleEndian.PutUint64(itemBytes[j*elemLen:], v)
						}

					case reflect.Float32:
						item := fieldValue.Index(i).Interface().([]float32)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							bits := math.Float32bits(v)
							binary.LittleEndian.PutUint32(itemBytes[j*elemLen:], bits)
						}
					case reflect.Float64:
						item := fieldValue.Index(i).Interface().([]float64)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							bits := math.Float64bits(v)
							binary.LittleEndian.PutUint64(itemBytes[j*elemLen:], bits)
						}

					default:
						return nil, fmt.Errorf("unsupported slice element type: %s", elemBaseKind)
					}

					arrayLen := array_elem_len * elemLen

					// Ensure proper length adjustments if needed
					if arrayLen > 0 {
						if len(itemBytes) > arrayLen {
							itemBytes = itemBytes[:arrayLen]
						} else if len(itemBytes) < arrayLen {
							itemBytes = append(itemBytes, make([]byte, arrayLen-len(itemBytes))...)
						}
					} else {
						itemLen := len(itemBytes) / elemLen
						writeByteLen(buf, numBytes, itemLen)
					}

					buf.Write(itemBytes)
				}
			} else if elemKind == reflect.Int8 || elemKind == reflect.Uint8 ||
				elemKind == reflect.Int16 || elemKind == reflect.Uint16 ||
				elemKind == reflect.Int32 || elemKind == reflect.Uint32 ||
				elemKind == reflect.Int64 || elemKind == reflect.Uint64 ||
				elemKind == reflect.Float64 || elemKind == reflect.Float32 {

				// Truncate using reflection
				truncatedSlice := fieldValue.Slice(0, length).Interface()

				// Write entire slice in one go
				binary.Write(buf, binary.LittleEndian, truncatedSlice)
			}
		default:
			return nil, fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}
	return buf.Bytes(), nil
}

// Deserialize binary []byte into struct (Little-Endian)
func Deserialize(data []byte, s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return errors.New("DeserializeBinary: expected a pointer")
	}

	buf := bytes.NewReader(data)
	val = val.Elem()
	typ := val.Type()

	sortedFields := getSortedFields(typ)
	for _, field := range sortedFields {

		tag := field.Tag.Get("bin")
		if tag == "" {
			continue
		}
		fieldValue, numBytes, _, _, array_elem_len := extractFieldParameters(val, field)

		// Handle fields based on their types
		switch fieldValue.Kind() {
		case reflect.Int8:
			var num int8
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetInt(int64(num))
		case reflect.Uint8:
			var num uint8
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetUint(uint64(num))
		case reflect.Int16:
			var num int16
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetInt(int64(num))
		case reflect.Uint16:
			var num uint16
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetUint(uint64(num))
		case reflect.Int32:
			var num int32
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetInt(int64(num))
		case reflect.Uint32:
			var num uint32
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetUint(uint64(num))
		case reflect.Int64:
			var num int64
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetInt(num)
		case reflect.Uint64:
			var num uint64
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetUint(num)
		case reflect.Float64:
			var num float64
			binary.Read(buf, binary.LittleEndian, &num)
			fieldValue.SetFloat(num)
		case reflect.String:
			length := readByteLen(buf, numBytes)
			strBytes := make([]byte, length)
			buf.Read(strBytes)
			fieldValue.SetString(string(strBytes))
		case reflect.Slice:
			elemKind := fieldValue.Type().Elem().Kind()
			elemType := fieldValue.Type().Elem()

			length := readByteLen(buf, numBytes)

			// if length == 0 { // Ensure nil is restored instead of empty slice
			// 	fieldValue.Set(reflect.Zero(fieldValue.Type()))
			// 	continue
			// }

			if elemKind == reflect.Uint8 { // Special case for []byte
				byteData := make([]byte, length)
				buf.Read(byteData)
				fieldValue.SetBytes(byteData)
			} else if elemKind == reflect.Slice {
				elemBaseKind := elemType.Elem().Kind()
				elemLen := reflectKindByteLen(elemBaseKind)

				// Create a new slice of the correct type and length
				newSlice := reflect.MakeSlice(fieldValue.Type(), length, length)

				for i := 0; i < length; i++ {
					var itemBytes []byte

					if array_elem_len > 0 {
						itemBytes = make([]byte, array_elem_len*elemLen)
					} else {
						itemLength := readByteLen(buf, numBytes) * elemLen
						itemBytes = make([]byte, itemLength)
					}

					// Read exactly array_elem_len bytes
					_, err := buf.Read(itemBytes)
					if err != nil {
						return fmt.Errorf("failed to read slice element: %w", err)
					}

					switch elemBaseKind {
					case reflect.Uint8, reflect.Int8:
						newSlice.Index(i).Set(reflect.ValueOf(itemBytes))

					case reflect.Int16:
						item := make([]int16, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = int16(binary.LittleEndian.Uint16(itemBytes[j*elemLen:]))
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Uint16:
						item := make([]uint16, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = binary.LittleEndian.Uint16(itemBytes[j*elemLen:])
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					case reflect.Int32:
						item := make([]int32, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = int32(binary.LittleEndian.Uint32(itemBytes[j*elemLen:]))
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Uint32:
						item := make([]uint32, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = binary.LittleEndian.Uint32(itemBytes[j*elemLen:])
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					case reflect.Int64:
						item := make([]int64, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = int64(binary.LittleEndian.Uint64(itemBytes[j*elemLen:]))
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Uint64:
						item := make([]uint64, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = binary.LittleEndian.Uint64(itemBytes[j*elemLen:])
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					case reflect.Float32:
						item := make([]float32, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							bits := binary.LittleEndian.Uint32(itemBytes[j*elemLen:])
							item[j] = math.Float32frombits(bits)
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Float64:
						item := make([]float64, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							bits := binary.LittleEndian.Uint64(itemBytes[j*elemLen:])
							item[j] = math.Float64frombits(bits)
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					default:
						return fmt.Errorf("unsupported slice element type: %s", elemBaseKind)
					}
				}

				// Set the deserialized slice to the field
				fieldValue.Set(newSlice)
			} else if elemKind == reflect.Int8 || elemKind == reflect.Uint8 ||
				elemKind == reflect.Int16 || elemKind == reflect.Uint16 ||
				elemKind == reflect.Int32 || elemKind == reflect.Uint32 ||
				elemKind == reflect.Int64 || elemKind == reflect.Uint64 ||
				elemKind == reflect.Float64 || elemKind == reflect.Float32 {

				// Create a new slice of the correct type and length
				newSlice := reflect.MakeSlice(fieldValue.Type(), length, length)

				// Read the entire slice in one go
				err := binary.Read(buf, binary.LittleEndian, newSlice.Interface())
				if err != nil {
					return fmt.Errorf("failed to read slice: %w", err)
				}

				// Set the field with the new slice
				fieldValue.Set(newSlice)
			}
		default:
			return fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}
	return nil
}

// Compare compares two structs field by field (Little-Endian style)
func Compare(a, b interface{}) (bool, error) {
	// Ensure we're working with non-pointer values
	if reflect.ValueOf(a).Kind() == reflect.Ptr || reflect.ValueOf(b).Kind() == reflect.Ptr {
		return false, errors.New("CompareStruct: expected non-pointer values")
	}

	// Ensure both structs are of the same type
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false, errors.New("CompareStruct: mismatched types")
	}

	// Get reflection values
	valA := reflect.ValueOf(a)
	valB := reflect.ValueOf(b)
	typ := reflect.TypeOf(a)

	sortedFields := getSortedFields(typ)
	for i := range sortedFields {

		fieldA := valA.Field(i)
		fieldB := valB.Field(i)

		tag := typ.Field(i).Tag.Get("bin")

		// Skip fields without bin tag
		if tag == "" {
			continue
		}

		// Check if the field types match
		if fieldA.Kind() != fieldB.Kind() {
			return false, nil // Field types don't match
		}

		// Compare values based on field type
		switch fieldA.Kind() {
		case reflect.Int8:
			if fieldA.Int() != fieldB.Int() {
				return false, nil
			}
		case reflect.Uint8:
			if fieldA.Uint() != fieldB.Uint() {
				return false, nil
			}
		case reflect.Int16:
			if fieldA.Int() != fieldB.Int() {
				return false, nil
			}
		case reflect.Uint16:
			if fieldA.Uint() != fieldB.Uint() {
				return false, nil
			}
		case reflect.Int32:
			if fieldA.Int() != fieldB.Int() {
				return false, nil
			}
		case reflect.Uint32:
			if fieldA.Uint() != fieldB.Uint() {
				return false, nil
			}
		case reflect.Int64:
			if fieldA.Int() != fieldB.Int() {
				return false, nil
			}
		case reflect.Uint64:
			if fieldA.Uint() != fieldB.Uint() {
				return false, nil
			}
		case reflect.Float32:
			if fieldA.Float() != fieldB.Float() {
				return false, nil
			}
		case reflect.Float64:
			if fieldA.Float() != fieldB.Float() {
				return false, nil
			}
		case reflect.String:
			if fieldA.String() != fieldB.String() {
				return false, nil
			}
		case reflect.Slice:
			if fieldA.Len() != fieldB.Len() {
				return false, nil
			}
			for j := 0; j < fieldA.Len(); j++ {
				elemA := fieldA.Index(j).Interface()
				elemB := fieldB.Index(j).Interface()

				// Compare primitive elements directly
				if !reflect.DeepEqual(elemA, elemB) {
					return false, nil
				}
			}
		default:
			// If a field type is unsupported, return an error
			return false, fmt.Errorf("unsupported type: %s", fieldA.Kind())
		}
	}

	return true, nil
}

func MarshalJSON(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr && val.Kind() != reflect.Struct {
		return nil, errors.New("MarshalJSON: expected a struct or pointer to struct")
	}

	val = reflect.Indirect(val)
	typ := val.Type()
	jsonMap := make(map[string]interface{})

	sortedFields := getSortedFields(typ)
	for i, field := range sortedFields {

		tag := field.Tag.Get("bin")
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue := val.Field(i)
		switch fieldValue.Kind() {
		case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			jsonMap[tag] = fieldValue.Int()
		case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			jsonMap[tag] = fieldValue.Uint()
		case reflect.Float64, reflect.Float32:
			jsonMap[tag] = fieldValue.Float()
		case reflect.String:
			jsonMap[tag] = fieldValue.String()
		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() == reflect.Uint8 { // Handle []byte
				jsonMap[tag] = base64.StdEncoding.EncodeToString(fieldValue.Bytes())
			} else {
				if fieldValue.IsNil() {
					jsonMap[tag] = reflect.MakeSlice(fieldValue.Type(), 0, 0).Interface() // Ensure []
				} else {
					jsonMap[tag] = fieldValue.Interface()
				}
			}

		default:
			return nil, fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}

	return json.Marshal(jsonMap)
}

func getArrayElemLenFromField(field reflect.StructField) int {
	tag := field.Tag.Get("array_elem_len")

	var size int
	_, err := fmt.Sscanf(tag, "%d", &size)
	if err != nil {
		return 0
	}
	return size
}

func getSizeFromField(field reflect.StructField) int {
	tag := field.Tag.Get("max")

	var size int
	_, err := fmt.Sscanf(tag, "%d", &size)
	if err != nil {
		return 0
	}
	return size
}

// Get the bit from struct tag (default to 8 if not specified)
func getByteFromField(field reflect.StructField) int {
	tag := field.Tag.Get("byte")

	var byte int
	_, err := fmt.Sscanf(tag, "%d", &byte)
	if err != nil && byte != 2 && byte != 3 && byte != 4 {
		return 1
	}
	return byte
}

func extractFieldParameters(val reflect.Value, field reflect.StructField) (reflect.Value, int, int, int, int) {
	// Get field value using FieldByIndex
	fieldValue := val.FieldByIndex(field.Index)

	numBytes := getByteFromField(field)
	maxSize := 1<<(numBytes*8) - 1
	size := getSizeFromField(field)
	if size == 0 {
		size = maxSize
	}

	arrayElemLen := getArrayElemLenFromField(field)

	return fieldValue, numBytes, maxSize, size, arrayElemLen
}

func writeByteLen(buf *bytes.Buffer, numByte int, length int) {
	switch numByte {
	case 2:
		binary.Write(buf, binary.LittleEndian, uint16(length))
	case 3:
		binary.Write(buf, binary.LittleEndian, uint32(length))
	case 4:
		binary.Write(buf, binary.LittleEndian, uint64(length))
	default:
		binary.Write(buf, binary.LittleEndian, uint8(length))
	}
}

func readByteLen(buf *bytes.Reader, numByte int) int {
	var length int
	switch numByte {
	case 2:
		var temp uint16
		binary.Read(buf, binary.LittleEndian, &temp)
		length = int(temp)
	case 3:
		var temp uint32
		binary.Read(buf, binary.LittleEndian, &temp)
		length = int(temp)
	case 4:
		var temp uint64
		binary.Read(buf, binary.LittleEndian, &temp)
		length = int(temp)
	default:
		var temp uint8
		binary.Read(buf, binary.LittleEndian, &temp)
		length = int(temp) // Default to 8-bit length
	}

	return length
}

func reflectKindByteLen(elemBaseKind reflect.Kind) int {
	switch elemBaseKind {
	case reflect.Uint16, reflect.Int16:
		return 2
	case reflect.Uint32, reflect.Int32, reflect.Float32:
		return 4
	case reflect.Uint64, reflect.Int64, reflect.Float64:
		return 8
	default:
		return 1
	}
}

func getSortedFields(typ reflect.Type) []reflect.StructField {
	numFields := typ.NumField()
	fields := make([]reflect.StructField, numFields)

	for i := 0; i < numFields; i++ {
		fields[i] = typ.Field(i)
	}

	// Sort fields by "bin" tag value (convert to int for correct order)
	sort.Slice(fields, func(i, j int) bool {
		tagI := fields[i].Tag.Get("bin")
		tagJ := fields[j].Tag.Get("bin")
		return tagI < tagJ
	})

	return fields
}

func hash(data interface{}) (string, error) {
	// Serialize the struct to JSON
	serialized, err := MarshalJSON(data)
	if err != nil {
		return "", err
	}

	// Compute MD5 hash of the serialized data
	hash := md5.New()
	hash.Write(serialized)

	// Get the hash sum as a byte slice
	hashBytes := hash.Sum(nil)

	// Return the hash as a hexadecimal string
	return hex.EncodeToString(hashBytes), nil
}
