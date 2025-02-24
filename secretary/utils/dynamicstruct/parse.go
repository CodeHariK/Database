package dynamicstruct

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type FieldInfo struct {
	Type string `json:"type"`
	Enc  int    `json:"enc,omitempty"`
	Size int    `json:"size,omitempty"`
}

func Parse(jsonData string) {
	// Parse the JSON metadata
	var metadata map[string]FieldInfo
	err := json.Unmarshal([]byte(jsonData), &metadata)
	if err != nil {
		panic(err)
	}

	// Create a dynamic struct based on metadata
	fields := []reflect.StructField{}
	for name, info := range metadata {
		var fieldType reflect.Type
		switch info.Type {
		case "int":
			fieldType = reflect.TypeOf(0)
		case "string":
			fieldType = reflect.TypeOf("")
		default:
			continue
		}

		// Capitalize field name
		fieldName := strings.ToUpper(name[:1]) + name[1:]

		fields = append(fields, reflect.StructField{
			Name: fieldName,
			Type: fieldType,
			Tag:  reflect.StructTag(fmt.Sprintf(`custom:"enc=%d,size=%d"`, info.Enc, info.Size)),
		})
	}

	// Create a struct type dynamically
	dynamicType := reflect.StructOf(fields)
	instance := reflect.New(dynamicType).Elem()

	fmt.Println(dynamicType)
	fmt.Println(instance)

	// Assign values dynamically
	if len(fields) > 0 && fields[0].Type.Kind() == reflect.Int {
		instance.Field(0).SetInt(42) // Set int value
	}
	if len(fields) > 1 && fields[1].Type.Kind() == reflect.String {
		instance.Field(1).SetString("abc") // Set string value
	}

	// Print dynamic struct values
	fmt.Println("Dynamically Created Struct:")
	for i := 0; i < dynamicType.NumField(); i++ {
		field := dynamicType.Field(i)
		fmt.Printf("%s: %v\n", field.Name, instance.Field(i).Interface())
	}

	// Access struct tags
	fmt.Println("\nStruct Field Tags:")
	for i := 0; i < dynamicType.NumField(); i++ {
		field := dynamicType.Field(i)
		fmt.Printf("%s tag: %s\n", field.Name, field.Tag)
	}

	fmt.Println(instance)
}
