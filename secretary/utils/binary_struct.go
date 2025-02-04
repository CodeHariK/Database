package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

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

func extractFieldParameters(val reflect.Value, i int, field reflect.StructField) (reflect.Value, int, int, int) {
	fieldValue := val.Field(i)
	numBytes := getByteFromField(field)
	maxSize := 1<<(numBytes*8) - 1
	size := getSizeFromField(field)
	if size == 0 {
		size = maxSize
	}

	// fmt.Printf("\nnumBytes:%d maxSize:%d size:%d\n", numBytes, maxSize, size)

	return fieldValue, numBytes, maxSize, size
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

// Serialize struct to binary []byte (Little-Endian)
// bin : type name
// byte : number of bytes used for length of string or []byte
// max : max length of string or []byte
func BinaryStructSerialize(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		return nil, errors.New("SerializeBinary: expected a value not pointer")
	}

	buf := new(bytes.Buffer)
	typ := reflect.TypeOf(s)

	for i := 0; i < typ.NumField(); i++ {

		field := typ.Field(i)
		tag := field.Tag.Get("bin")

		// Skip fields without bin tag
		if tag == "" {
			continue
		}
		fieldValue, numBytes, maxStorableSize, maxSize := extractFieldParameters(val, i, field)

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
			} else if elemKind == reflect.Int8 || elemKind == reflect.Uint8 ||
				elemKind == reflect.Int16 || elemKind == reflect.Uint16 ||
				elemKind == reflect.Int32 || elemKind == reflect.Uint32 ||
				elemKind == reflect.Int64 || elemKind == reflect.Uint64 ||
				elemKind == reflect.Float64 {

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

// BinaryStructDeserialize binary []byte into struct (Little-Endian)
func BinaryStructDeserialize(data []byte, s interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr {
		return errors.New("DeserializeBinary: expected a pointer")
	}

	buf := bytes.NewReader(data)
	val = val.Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {

		field := typ.Field(i)
		tag := field.Tag.Get("bin")
		if tag == "" {
			continue
		}
		fieldValue, numBytes, _, _ := extractFieldParameters(val, i, field)

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
			length := readByteLen(buf, numBytes)

			if length == 0 { // Ensure nil is restored instead of empty slice
				fieldValue.Set(reflect.Zero(fieldValue.Type()))
				continue
			}

			if elemKind == reflect.Uint8 { // Special case for []byte
				byteData := make([]byte, length)
				buf.Read(byteData)
				fieldValue.SetBytes(byteData)
			} else if elemKind == reflect.Int8 || elemKind == reflect.Uint8 ||
				elemKind == reflect.Int16 || elemKind == reflect.Uint16 ||
				elemKind == reflect.Int32 || elemKind == reflect.Uint32 ||
				elemKind == reflect.Int64 || elemKind == reflect.Uint64 ||
				elemKind == reflect.Float64 {

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

// BinaryStructCompare compares two structs field by field (Little-Endian style)
func BinaryStructCompare(a, b interface{}) (bool, error) {
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

	for i := 0; i < typ.NumField(); i++ {
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
				if fieldA.Index(j).Interface() != fieldB.Index(j).Interface() {
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

func BinaryStructMarshalJSON(s interface{}) ([]byte, error) {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr && val.Kind() != reflect.Struct {
		return nil, errors.New("MarshalJSON: expected a struct or pointer to struct")
	}

	val = reflect.Indirect(val)
	typ := val.Type()
	jsonMap := make(map[string]interface{})

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
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
				jsonMap[tag] = fieldValue.Interface()
			}
		default:
			return nil, fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}

	return json.Marshal(jsonMap)
}
