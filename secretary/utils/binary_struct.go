package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
)

func getSizeFromField(field reflect.StructField) int {
	tag := field.Tag.Get("size")

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
func SerializeBinary(s interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	val := reflect.ValueOf(s)
	typ := reflect.TypeOf(s)

	for i := 0; i < typ.NumField(); i++ {

		field := typ.Field(i)
		tag := field.Tag.Get("bin")
		if tag == "" {
			continue
		}
		fieldValue, numBytes, maxSize, size := extractFieldParameters(val, i, field)

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
			if len(str) >= size {
				str = str[:size]
			}
			if len(str) >= maxSize {
				str = str[:maxSize]
			}
			writeByteLen(buf, []byte(str), numBytes)
			buf.WriteString(str) // Write string directly
		case reflect.Slice:
			if fieldValue.Type().Elem().Kind() == reflect.Uint8 { // Handle []byte
				data := fieldValue.Bytes()
				if len(data) >= size {
					data = data[:size]
				}
				if len(data) >= maxSize {
					data = data[:maxSize]
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

// Deserialize binary []byte into struct (Little-Endian)
func DeserializeBinary(data []byte, s interface{}) error {
	buf := bytes.NewReader(data)
	val := reflect.ValueOf(s).Elem()
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
