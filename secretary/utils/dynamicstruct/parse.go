package dynamicstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/codeharik/secretary/utils"
)

// type FieldInfo struct {
// 	Type string `json:"type"`
// 	Enc  int    `json:"enc,omitempty"`
// 	Size int    `json:"size,omitempty"`
// }

// func Parse(jsonData string) {
// 	// Parse the JSON metadata
// 	var metadata map[string]FieldInfo
// 	err := json.Unmarshal([]byte(jsonData), &metadata)
// 	if err != nil {
// 		panic(err)
// 	}

// 	// Create a dynamic struct based on metadata
// 	fields := []reflect.StructField{}
// 	for name, info := range metadata {
// 		var fieldType reflect.Type
// 		switch info.Type {
// 		case "int":
// 			fieldType = reflect.TypeOf(0)
// 		case "string":
// 			fieldType = reflect.TypeOf("")
// 		default:
// 			continue
// 		}

// 		// Capitalize field name
// 		fieldName := strings.ToUpper(name[:1]) + name[1:]

// 		fields = append(fields, reflect.StructField{
// 			Name: fieldName,
// 			Type: fieldType,
// 			Tag:  reflect.StructTag(fmt.Sprintf(`custom:"enc=%d,size=%d"`, info.Enc, info.Size)),
// 		})
// 	}

// 	// Create a struct type dynamically
// 	dynamicType := reflect.StructOf(fields)
// 	instance := reflect.New(dynamicType).Elem()

// 	fmt.Println(dynamicType)
// 	fmt.Println(instance)

// 	// Assign values dynamically
// 	if len(fields) > 0 && fields[0].Type.Kind() == reflect.Int {
// 		instance.Field(0).SetInt(42) // Set int value
// 	}
// 	if len(fields) > 1 && fields[1].Type.Kind() == reflect.String {
// 		instance.Field(1).SetString("abc") // Set string value
// 	}

// 	// Print dynamic struct values
// 	fmt.Println("Dynamically Created Struct:")
// 	for i := 0; i < dynamicType.NumField(); i++ {
// 		field := dynamicType.Field(i)
// 		fmt.Printf("%s: %v\n", field.Name, instance.Field(i).Interface())
// 	}

// 	// Access struct tags
// 	fmt.Println("\nStruct Field Tags:")
// 	for i := 0; i < dynamicType.NumField(); i++ {
// 		field := dynamicType.Field(i)
// 		fmt.Printf("%s tag: %s\n", field.Name, field.Tag)
// 	}

// 	fmt.Println(instance)
// }
// FieldSchema represents a field's metadata
// It dynamically stores all tags in a map

type FieldSchema struct {
	Type string            `json:"type"`
	Tags map[string]string `json:"tags,omitempty"`
}

// Convert a struct to a JSON schema (Handles arbitrary tags)
func StructToSchema(v interface{}) (string, error) {
	t := reflect.TypeOf(v)
	schema := make(map[string]FieldSchema)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldType := ""

		// Determine Go type to JSON type
		switch field.Type.Kind() {
		case reflect.Int:
			fieldType = "int"
		case reflect.String:
			fieldType = "string"
		case reflect.Float64:
			fieldType = "float"
		case reflect.Bool:
			fieldType = "bool"
		default:
			continue
		}

		// Extract all tags dynamically
		tagMap := make(map[string]string)
		for _, tag := range strings.Split(string(field.Tag), " ") {
			parts := strings.SplitN(tag, ":", 2)
			if len(parts) == 2 {
				tagMap[strings.Trim(parts[0], "` ")] = strings.Trim(parts[1], "`\"")
			}
		}

		// Convert to schema format
		fieldSchema := FieldSchema{
			Type: fieldType,
			Tags: tagMap,
		}

		schema[field.Name] = fieldSchema
	}

	// Convert schema map to JSON
	jsonSchema, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonSchema), nil
}

// Convert JSON schema back to a dynamic struct
type DynamicStruct struct {
	Type     reflect.Type
	Instance reflect.Value
}

func SchemaToStruct(schemaJSON string) (*DynamicStruct, error) {
	var schema map[string]FieldSchema
	err := json.Unmarshal([]byte(schemaJSON), &schema)
	if err != nil {
		return nil, err
	}

	fields := []reflect.StructField{}

	for name, info := range schema {
		var fieldType reflect.Type
		switch info.Type {
		case "int":
			fieldType = reflect.TypeOf(0)
		case "string":
			fieldType = reflect.TypeOf("")
		case "float":
			fieldType = reflect.TypeOf(0.0)
		case "bool":
			fieldType = reflect.TypeOf(true)
		default:
			continue
		}

		// Construct struct tag correctly
		tagParts := []string{}
		for k, v := range info.Tags {
			tagParts = append(tagParts, fmt.Sprintf(`%s:"%s"`, k, v))
		}
		tagStr := strings.Join(tagParts, " ")

		fields = append(fields, reflect.StructField{
			Name: name,
			Type: fieldType,
			Tag:  reflect.StructTag(tagStr),
		})
	}

	dynamicType := reflect.StructOf(fields)
	return &DynamicStruct{
		Type:     dynamicType,
		Instance: reflect.New(dynamicType).Elem(),
	}, nil
}

func (ds *DynamicStruct) NewInstance() *DynamicStruct {
	return &DynamicStruct{
		Type:     ds.Type,
		Instance: reflect.New(ds.Type).Elem(),
	}
}

func (ds *DynamicStruct) SetField(name string, value any) error {
	field := ds.Instance.FieldByName(name)

	// Check if the field exists
	if !field.IsValid() {
		return fmt.Errorf("field %s not found in struct", name)
	}

	// Ensure the field is settable
	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", name)
	}

	// Handle different types dynamically
	switch field.Kind() {
	case reflect.String:
		strValue, ok := value.(string)
		if !ok {
			return fmt.Errorf("field %s expects a string, got %T", name, value)
		}
		field.SetString(strValue)

	case reflect.Float64:
		floatValue, ok := value.(float64)
		if !ok {
			return fmt.Errorf("field %s expects a float64, got %T", name, value)
		}
		field.SetFloat(floatValue)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, ok := value.(int64)
		if !ok {
			return fmt.Errorf("field %s expects an int64, got %T", name, value)
		}
		field.SetInt(intValue)

	case reflect.Bool:
		boolValue, ok := value.(bool)
		if !ok {
			return fmt.Errorf("field %s expects a bool, got %T", name, value)
		}
		field.SetBool(boolValue)

	default:
		return fmt.Errorf("unsupported field type %s for %s", field.Kind(), name)
	}

	return nil
}

func (ds *DynamicStruct) ValidateStruct(instance interface{}) error {
	v := reflect.ValueOf(instance)

	for i := 0; i < ds.Type.NumField(); i++ {
		field := ds.Type.Field(i)
		value := v.FieldByName(field.Name)

		regexTag := field.Tag.Get("regex")
		errorMsg := strings.ReplaceAll(field.Tag.Get("error"), "_", " ")

		utils.Log("index", i,
			"value", fmt.Sprint(value),
			"regexTag", regexTag,
			"v", v.Type().Name(),
			"errorTag", errorMsg,
			"field.Type", field.Type,
			"field.Name", field.Name,
			"field.Tag", field.Tag)

		if regexTag != "" {
			if !utils.ValidateRegex(fmt.Sprint(value), regexTag) {
				if errorMsg != "" {
					return errors.New(errorMsg)
				}
				return fmt.Errorf("field %s does not match regex %s", field.Name, regexTag)
			}
		}
	}
	return nil
}

func (ds *DynamicStruct) ToSchema() (string, error) {
	instance := reflect.New(ds.Type).Elem().Interface()

	return StructToSchema(instance)
}

// Convert Dynamic Struct to Real Struct
// func (ds *DynamicType) DynamicToRealStruct(realStruct any) error {
// 	realValue := reflect.ValueOf(realStruct)
// 	if realValue.Kind() != reflect.Ptr || realValue.Elem().Kind() != reflect.Struct {
// 		return fmt.Errorf("realStruct must be a pointer to a struct")
// 	}

// 	realValue = realValue.Elem() // Dereference pointer to struct

// 	dynamicValue := reflect.ValueOf(ds.Type)
// 	if dynamicValue.Kind() == reflect.Ptr {
// 		dynamicValue = reflect.ValueOf(dynamicValue.Elem()) // Dereference if it's a pointer
// 	}

// 	for i := 0; i < dynamicValue.NumField(); i++ {
// 		field := ds.Type.Field(i)
// 		fieldValue := dynamicValue.Field(i) // Get actual field value
// 		fieldName := field.Name

// 		v, _ := ds.Type.FieldByName(field.Name)

// 		utils.Log(
// 			"fieldValue", fieldValue,
// 			"fieldValue.Type().Name()", fieldValue.Type().Name(),
// 			"field.Name", field.Name)

// 		realField := realValue.FieldByName(fieldName)
// 		if !realField.IsValid() {
// 			return fmt.Errorf("field %s not found in real struct", fieldName)
// 		}
// 		if !realField.CanSet() {
// 			return fmt.Errorf("field %s cannot be set", fieldName)
// 		}

// 		utils.Log("fieldValue.Type().Name()", fieldValue.Type().Name(),
// 			"fieldValue.Type().Kind()", fieldValue.Type().Kind(),
// 			"fieldValue.Type().String()", fieldValue.Type().String(),
// 			"fieldValue", fieldValue,
// 			"fieldValue", fieldValue.String(),
// 			"realField.Type()", realField.Type(),
// 			"v", fmt.Sprint(v),
// 			"fieldValue.Type()", fieldValue.Type(),
// 			"fieldValue.Kind()", fieldValue.Kind(),
// 			"realField.Type()", realField.Type(),
// 			"fieldValue.Elem().Type()", fieldValue.Elem().Type())

// 		if realField.Type() != fieldValue.Type() {
// 			// If fieldValue is a pointer, try to get the underlying value
// 			if fieldValue.Kind() == reflect.Ptr && fieldValue.Elem().Type() == realField.Type() {
// 				fieldValue = fieldValue.Elem() // Dereference pointer
// 			} else {
// 				return fmt.Errorf("type mismatch for field %s: expected %s, got %s", fieldName, realField.Type(), fieldValue.Type())
// 			}
// 		}

// 		realField.Set(fieldValue) // Correctly set the value
// 	}

// 	return nil
// }

func (ds *DynamicStruct) DynamicToRealStruct(realStruct interface{}) error {
	dynValue := reflect.ValueOf(ds.Instance.Interface())
	realValue := reflect.ValueOf(realStruct).Elem()

	// Ensure both are structs
	if dynValue.Kind() != reflect.Struct || realValue.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct types")
	}

	// Copy values by field name
	for i := 0; i < realValue.NumField(); i++ {
		realField := realValue.Type().Field(i)           // Get field metadata
		dynField := dynValue.FieldByName(realField.Name) // Get field by name from dynamic struct

		if dynField.IsValid() && dynField.Type() == realField.Type {
			realValue.Field(i).Set(dynField) // Set value if types match
		} else if dynField.IsValid() {
			return fmt.Errorf("type mismatch for field %s: expected %s, got %s",
				realField.Name, realField.Type, dynField.Type())
		} else {
			return fmt.Errorf("field %s not found in dynamic struct", realField.Name)
		}
	}

	return nil
}
