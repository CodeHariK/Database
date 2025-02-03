package utils

import (
	"bytes"
	"encoding/binary"
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

func writeByteLen(buf *bytes.Buffer, data []byte, numByte int) {
	switch numByte {
	case 2:
		binary.Write(buf, binary.LittleEndian, uint16(len(data)))
	case 3:
		binary.Write(buf, binary.LittleEndian, uint32(len(data)))
	case 4:
		binary.Write(buf, binary.LittleEndian, uint64(len(data)))
	default:
		binary.Write(buf, binary.LittleEndian, uint8(len(data)))
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
func SerializeBinaryStruct(s interface{}) ([]byte, error) {
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
			writeByteLen(buf, []byte(str), numBytes)
			buf.WriteString(str) // Write string directly
		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() == reflect.Uint8 { // Handle []byte
				data := fieldValue.Bytes()
				if len(data) >= maxSize {
					data = data[:maxSize]
				}
				if len(data) >= maxStorableSize {
					data = data[:maxStorableSize]
				}

				writeByteLen(buf, data, numBytes)
				buf.Write(data) // Write  directly
			}
		default:
			return nil, fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}
	return buf.Bytes(), nil
}

// Deserialize binary []byte into struct (Little-Endian)
func DeserializeBinaryStruct(data []byte, s interface{}) error {
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
			if fieldValue.Type().Elem().Kind() == reflect.Uint8 { // Handle []byte
				length := readByteLen(buf, numBytes)

				if length == 0 { // Ensure nil is restored instead of empty slice
					fieldValue.Set(reflect.Zero(fieldValue.Type()))
					continue
				}

				byteData := make([]byte, length)
				buf.Read(byteData)
				fieldValue.SetBytes(byteData)
			}
		default:
			return fmt.Errorf("unsupported type: %s", field.Type.Kind())
		}
	}
	return nil
}

// CompareBinaryStruct compares two structs field by field (Little-Endian style)
func CompareBinaryStruct(a, b interface{}) (bool, error) {
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
