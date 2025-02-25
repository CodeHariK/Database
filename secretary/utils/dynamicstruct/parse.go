package dynamicstruct

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/codeharik/secretary/utils"
)

type FieldSchema struct {
	Type string            `json:"type"`
	Tags map[string]string `json:"tags,omitempty"`
}

type DynamicStruct struct {
	Type     reflect.Type
	Instance reflect.Value
}

// Convert a struct to a JSON schema (Handles arbitrary tags)
func StructToSchema(instance any) (string, error) {
	t := reflect.TypeOf(instance)
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

func (ds *DynamicStruct) ToSchema() (string, error) {
	instance := reflect.New(ds.Type).Elem().Interface()

	return StructToSchema(instance)
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

func (ds *DynamicStruct) Validate() error {
	return ds.ValidateStruct(ds.Instance.Interface())
}

func (ds *DynamicStruct) ValidateStruct(instance any) error {
	ivalue := reflect.ValueOf(instance)

	for i := 0; i < ds.Type.NumField(); i++ {
		field := ds.Type.Field(i)
		value := ivalue.FieldByName(field.Name)

		regexTag := field.Tag.Get("regex")
		errorMsg := strings.ReplaceAll(field.Tag.Get("error"), "_", " ")

		if regexTag != "" {
			if !utils.ValidateRegex(fmt.Sprint(value), regexTag) {
				if errorMsg != "" {
					return errors.New(errorMsg)
				}
				return fmt.Errorf("field %s, value %q does not match regex %s", field.Name, fmt.Sprint(value), regexTag)
			}
		}
	}
	return nil
}

func (ds *DynamicStruct) DynamicToRealStruct(realStruct any) error {
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
		rv := realValue.FieldByName(realField.Name)

		if dynField.IsValid() && rv.CanSet() && dynField.Type() == rv.Type() {
			rv.Set(dynField) // Set value if types match
		} else if dynField.IsValid() {
			return fmt.Errorf("type mismatch for field %s: expected %s, got %s",
				realField.Name, realField.Type, dynField.Type())
		} else {
			return fmt.Errorf("field %s not found in dynamic struct", realField.Name)
		}
	}

	return nil
}

func (ds *DynamicStruct) JsonUnmarshal(data []byte) error {
	err := json.Unmarshal(data, ds.Instance.Addr().Interface())
	if err != nil {
		fmt.Println("Unmarshal error:", err)
	}
	return err
}

func (dynamicMachine *DynamicStruct) JsonMarshal() ([]byte, error) {
	return json.Marshal(dynamicMachine.Instance.Interface())
}
