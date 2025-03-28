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
	"io"
	"math"
	"reflect"
	"sort"
)

var BYTEORDER binary.ByteOrder = binary.LittleEndian

// Serialize struct to binary []byte (Little-Endian)
// bin : type name
// lenbyte : number of bytes used for length of string or []byte
// max : max length of string or []byte
// array_elem_len : max length (array elements) in (array of (array elements)), [][]byte [][]int32 [][]float64
func Serialize(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference pointer
	}

	typ := val.Type()
	buf := new(bytes.Buffer)

	// If `s` is a slice of structs, serialize the slice length first
	if typ.Kind() == reflect.Slice {
		elemType := typ.Elem()
		if elemType.Kind() != reflect.Struct {
			return nil, errors.New("SerializeBinary: expected slice of structs")
		}

		// Write slice length
		sliceLen := val.Len()
		if err := binary.Write(buf, BYTEORDER, int32(sliceLen)); err != nil {
			return nil, err
		}

		// Serialize each struct in the slice
		for i := 0; i < sliceLen; i++ {

			// Serialize the struct into structBuf
			structData, err := Serialize(val.Index(i).Interface())
			if err != nil {
				return nil, err
			}

			// Write struct size before writing struct data
			structSize := int32(len(structData))
			if err := binary.Write(buf, BYTEORDER, structSize); err != nil {
				return nil, err
			}

			buf.Write(structData)
		}

		return buf.Bytes(), nil
	}

	sortedFields, err := getSortedFields(typ, true)
	if err != nil {
		return nil, err
	}

	for _, field := range sortedFields {

		tag := field.Tag.Get("bin")
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue, numBytes, maxStorableSize, maxSize, array_elem_len, err := extractFieldParameters(val, field)
		if err != nil {
			continue
		}

		// Handle fields based on their types
		switch fieldValue.Kind() {
		case reflect.Int8:
			binary.Write(buf, BYTEORDER, int8(fieldValue.Int()))
		case reflect.Uint8:
			binary.Write(buf, BYTEORDER, uint8(fieldValue.Uint()))
		case reflect.Int16:
			binary.Write(buf, BYTEORDER, int16(fieldValue.Int()))
		case reflect.Uint16:
			binary.Write(buf, BYTEORDER, uint16(fieldValue.Uint()))
		case reflect.Int32:
			binary.Write(buf, BYTEORDER, int32(fieldValue.Int()))
		case reflect.Uint32:
			binary.Write(buf, BYTEORDER, uint32(fieldValue.Uint()))
		case reflect.Int64:
			binary.Write(buf, BYTEORDER, int64(fieldValue.Int()))
		case reflect.Uint64:
			binary.Write(buf, BYTEORDER, uint64(fieldValue.Uint()))
		case reflect.Float64:
			binary.Write(buf, BYTEORDER, float64(fieldValue.Float()))
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

		case reflect.Struct:
			// Recursively serialize the nested struct
			structBytes, err := Serialize(fieldValue.Interface())
			if err != nil {
				continue
			}

			// Write struct size before the actual struct data
			structSize := int32(len(structBytes))
			if err := binary.Write(buf, BYTEORDER, structSize); err != nil {
				continue
			}

			// Write serialized struct to buffer
			_, err = buf.Write(structBytes)
			if err != nil {
				continue
			}

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
							BYTEORDER.PutUint16(itemBytes[j*elemLen:], uint16(v))
						}
					case reflect.Uint16:
						item := fieldValue.Index(i).Interface().([]uint16)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							BYTEORDER.PutUint16(itemBytes[j*elemLen:], v)
						}

					case reflect.Int32:
						item := fieldValue.Index(i).Interface().([]int32)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							BYTEORDER.PutUint32(itemBytes[j*elemLen:], uint32(v))
						}
					case reflect.Uint32:
						item := fieldValue.Index(i).Interface().([]uint32)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							BYTEORDER.PutUint32(itemBytes[j*elemLen:], v)
						}

					case reflect.Int64:
						item := fieldValue.Index(i).Interface().([]int64)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							BYTEORDER.PutUint64(itemBytes[j*elemLen:], uint64(v))
						}
					case reflect.Uint64:
						item := fieldValue.Index(i).Interface().([]uint64)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							BYTEORDER.PutUint64(itemBytes[j*elemLen:], v)
						}

					case reflect.Float32:
						item := fieldValue.Index(i).Interface().([]float32)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							bits := math.Float32bits(v)
							BYTEORDER.PutUint32(itemBytes[j*elemLen:], bits)
						}
					case reflect.Float64:
						item := fieldValue.Index(i).Interface().([]float64)
						itemBytes = make([]byte, len(item)*elemLen)
						for j, v := range item {
							bits := math.Float64bits(v)
							BYTEORDER.PutUint64(itemBytes[j*elemLen:], bits)
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
				binary.Write(buf, BYTEORDER, truncatedSlice)
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
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference pointer
	}

	typ := val.Type()
	buf := bytes.NewReader(data)

	// Handle slices of structs
	if typ.Kind() == reflect.Slice {
		elemType := typ.Elem()
		if elemType.Kind() != reflect.Struct {
			return errors.New("DeserializeBinary: expected slice of structs")
		}

		// Read the slice length
		var sliceLen int32
		if err := binary.Read(buf, BYTEORDER, &sliceLen); err != nil {
			return err
		}

		// Create a new slice with the required length
		slice := reflect.MakeSlice(typ, int(sliceLen), int(sliceLen))

		offset := int(binary.Size(sliceLen)) // Start tracking byte offset

		// Deserialize each struct using proper offset tracking
		for i := 0; i < int(sliceLen); i++ {
			// Read struct size
			var structSize int32
			if err := binary.Read(buf, BYTEORDER, &structSize); err != nil {
				return err
			}

			// Advance buf manually by reading struct bytes
			if _, err := buf.Seek(int64(structSize), io.SeekCurrent); err != nil {
				return err
			}

			offset += int(binary.Size(structSize)) // Advance offset for next struct

			// Extract correct struct data using offset
			structData := data[offset : offset+int(structSize)]

			offset += int(structSize) // Advance offset for next struct

			// Deserialize struct from the sliced data
			elem := reflect.New(elemType).Elem()
			if err := Deserialize(structData, elem.Addr().Interface()); err != nil {
				return err
			}

			slice.Index(i).Set(elem)
		}

		val.Set(slice) // Assign deserialized slice back
		return nil
	}

	sortedFields, err := getSortedFields(typ, true)
	if err != nil {
		return err
	}

	for _, field := range sortedFields {

		tag := field.Tag.Get("bin")
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue, numBytes, _, _, array_elem_len, err := extractFieldParameters(val, field)
		if err != nil {
			continue
		}

		// Handle fields based on their types
		switch fieldValue.Kind() {
		case reflect.Int8:
			var num int8
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetInt(int64(num))
		case reflect.Uint8:
			var num uint8
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetUint(uint64(num))
		case reflect.Int16:
			var num int16
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetInt(int64(num))
		case reflect.Uint16:
			var num uint16
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetUint(uint64(num))
		case reflect.Int32:
			var num int32
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetInt(int64(num))
		case reflect.Uint32:
			var num uint32
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetUint(uint64(num))
		case reflect.Int64:
			var num int64
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetInt(num)
		case reflect.Uint64:
			var num uint64
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetUint(num)
		case reflect.Float64:
			var num float64
			binary.Read(buf, BYTEORDER, &num)
			fieldValue.SetFloat(num)
		case reflect.String:
			length, err := readByteLen(buf, numBytes)
			if err != nil {
				continue
			}
			strBytes := make([]byte, length)
			buf.Read(strBytes)
			fieldValue.SetString(string(strBytes))

		case reflect.Struct:
			// Read struct size first
			var structSize int32
			if err := binary.Read(buf, BYTEORDER, &structSize); err != nil {
				continue
			}

			// Ensure structSize is valid
			if structSize <= 0 || structSize > int32(buf.Len()) {
				continue
			}

			// Read the struct bytes
			structBytes := make([]byte, structSize)
			if _, err := buf.Read(structBytes); err != nil {
				continue
			}

			// Recursively deserialize the nested struct
			if err := Deserialize(structBytes, fieldValue.Addr().Interface()); err != nil {
				continue
			}

		case reflect.Slice:
			elemKind := fieldValue.Type().Elem().Kind()
			elemType := fieldValue.Type().Elem()

			length, err := readByteLen(buf, numBytes)
			if err != nil {
				continue
			}

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
						itemLength, err := readByteLen(buf, numBytes)
						if err != nil {
							continue
						}
						itemLength *= elemLen
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
							item[j] = int16(BYTEORDER.Uint16(itemBytes[j*elemLen:]))
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Uint16:
						item := make([]uint16, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = BYTEORDER.Uint16(itemBytes[j*elemLen:])
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					case reflect.Int32:
						item := make([]int32, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = int32(BYTEORDER.Uint32(itemBytes[j*elemLen:]))
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Uint32:
						item := make([]uint32, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = BYTEORDER.Uint32(itemBytes[j*elemLen:])
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					case reflect.Int64:
						item := make([]int64, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = int64(BYTEORDER.Uint64(itemBytes[j*elemLen:]))
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Uint64:
						item := make([]uint64, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							item[j] = BYTEORDER.Uint64(itemBytes[j*elemLen:])
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))

					case reflect.Float32:
						item := make([]float32, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							bits := BYTEORDER.Uint32(itemBytes[j*elemLen:])
							item[j] = math.Float32frombits(bits)
						}
						newSlice.Index(i).Set(reflect.ValueOf(item))
					case reflect.Float64:
						item := make([]float64, len(itemBytes)/elemLen)
						for j := 0; j < len(item); j++ {
							bits := BYTEORDER.Uint64(itemBytes[j*elemLen:])
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
				err := binary.Read(buf, BYTEORDER, newSlice.Interface())
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

func MarshalJSON(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // Dereference pointer
	}

	typ := val.Type()
	jsonMap := make(map[string]interface{})

	if typ.Kind() == reflect.Slice {
		elemType := typ.Elem()
		if elemType.Kind() != reflect.Struct {
			return nil, errors.New("MarshalJSON: expected a struct or slice of structs")
		}

		// Handle slice of structs
		sliceLen := val.Len()
		jsonArray := make([]map[string]interface{}, sliceLen)

		for i := 0; i < sliceLen; i++ {
			itemJSON, err := MarshalJSON(val.Index(i).Interface()) // Recursive call
			if err != nil {
				return nil, err
			}

			var itemMap map[string]interface{}
			if err := json.Unmarshal(itemJSON, &itemMap); err != nil {
				return nil, err
			}
			jsonArray[i] = itemMap
		}

		return json.Marshal(jsonArray)
	}

	sortedFields, err := getSortedFields(typ, false)
	if err != nil {
		return nil, err
	}

	for _, field := range sortedFields {

		tag := field.Tag.Get("bin")
		if tag == "" || tag == "-" {
			continue
		}

		fieldValue, _, _, _, _, err := extractFieldParameters(val, field)
		if err != nil {
			continue
		}

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
			} else if fieldValue.Type().Elem().Kind() == reflect.Struct { // Handle []struct
				sliceLen := fieldValue.Len()
				structSlice := make([]map[string]interface{}, sliceLen)
				for i := 0; i < sliceLen; i++ {
					itemJSON, err := MarshalJSON(fieldValue.Index(i).Interface())
					if err != nil {
						return nil, err
					}
					var itemMap map[string]interface{}
					if err := json.Unmarshal(itemJSON, &itemMap); err != nil {
						return nil, err
					}
					structSlice[i] = itemMap
				}
				jsonMap[tag] = structSlice
			} else {
				if fieldValue.IsNil() {
					jsonMap[tag] = []interface{}{} // Ensure empty array []
				} else {
					jsonMap[tag] = fieldValue.Interface()
				}
			}
		case reflect.Struct: // Handle nested struct
			nestedJSON, err := MarshalJSON(fieldValue.Interface())
			if err != nil {
				return nil, err
			}
			var nestedMap map[string]interface{}
			if err := json.Unmarshal(nestedJSON, &nestedMap); err != nil {
				return nil, err
			}
			jsonMap[tag] = nestedMap
		default:
			return nil, fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}

	return json.Marshal(jsonMap)
}

func Compare(a, b interface{}) (bool, error) {
	hashA, errA := hash(a)
	hashB, errB := hash(b)
	if errA != nil || errB != nil {
		return false, errors.Join(errA, errB)
	}
	return hashA == hashB, nil
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
	tag := field.Tag.Get("lenbyte")

	var byte int
	_, err := fmt.Sscanf(tag, "%d", &byte)
	if err != nil || byte < 1 || byte > 8 {
		return 4
	}
	return byte
}

func extractFieldParameters(val reflect.Value, field reflect.StructField) (reflect.Value, int, int, int, int, error) {
	// If val is a slice, get its element type
	if val.Kind() == reflect.Slice {
		if val.Len() == 0 {
			return reflect.Value{}, 0, 0, 0, 0, errors.New("extractFieldParameters: cannot extract field from an empty slice")
		}
		val = val.Index(0) // Work with the first element
	}

	// Ensure it's a struct before accessing fields
	if val.Kind() != reflect.Struct {
		return reflect.Value{}, 0, 0, 0, 0, fmt.Errorf("extractFieldParameters: expected struct, got %s", val.Kind())
	}

	// Get field value using FieldByIndex
	fieldValue := val.FieldByIndex(field.Index)

	numBytes := getByteFromField(field)
	maxSize := 1<<(numBytes*8) - 1
	size := getSizeFromField(field)
	if size == 0 {
		size = maxSize
	}

	arrayElemLen := getArrayElemLenFromField(field)

	return fieldValue, numBytes, maxSize, size, arrayElemLen, nil
}

func writeByteLen(buf *bytes.Buffer, numByte int, length int) error {
	switch numByte {
	case 2:
		return binary.Write(buf, BYTEORDER, uint16(length))
	case 3:
		return binary.Write(buf, BYTEORDER, uint32(length))
	case 4:
		return binary.Write(buf, BYTEORDER, uint64(length))
	default:
		return binary.Write(buf, BYTEORDER, uint8(length))
	}
}

func readByteLen(buf *bytes.Reader, numByte int) (int, error) {
	var length int
	var err error
	switch numByte {
	case 2:
		var temp uint16
		err = binary.Read(buf, BYTEORDER, &temp)
		length = int(temp)
	case 3:
		var temp uint32
		err = binary.Read(buf, BYTEORDER, &temp)
		length = int(temp)
	case 4:
		var temp uint64
		err = binary.Read(buf, BYTEORDER, &temp)
		length = int(temp)
	default:
		var temp uint8
		err = binary.Read(buf, BYTEORDER, &temp)
		length = int(temp) // Default to 8-bit length
	}

	return length, err
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

func getSortedFields(typ reflect.Type, doSort bool) ([]reflect.StructField, error) {
	// If it's a slice, get the element type
	if typ.Kind() == reflect.Slice {
		typ = typ.Elem() // Extract struct type from slice
	}

	// Ensure it's a struct
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("getSortedFields: expected struct, got %s", typ.Kind())
	}

	numFields := typ.NumField()
	fields := make([]reflect.StructField, numFields)

	for i := 0; i < numFields; i++ {
		fields[i] = typ.Field(i)
	}

	if doSort {
		// Sort fields by "bin" tag value (convert to int for correct order)
		sort.Slice(fields, func(i, j int) bool {
			tagI := fields[i].Tag.Get("bin")
			tagJ := fields[j].Tag.Get("bin")
			return tagI < tagJ
		})
	}

	return fields, nil
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
