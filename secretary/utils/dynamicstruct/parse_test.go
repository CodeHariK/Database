package dynamicstruct

import (
	"fmt"
	"testing"
)

func TestParse2(t *testing.T) {
	schemaJSON := `{
		"Name": {"type": "string", "size": "3"},
		"Enc": {"type": "string", "enum": "base|ascii"},
		"Id": {"type": "float", "precision": "2"}
	}`

	// Convert JSON Schema → Dynamic Struct
	dynamicType, err := SchemaToStruct(schemaJSON)
	if err != nil {
		panic(err)
	}

	// Set dynamic struct values
	dynamicType.SetField("Name", "ABC")
	dynamicType.SetField("Enc", "ascii")
	dynamicType.SetField("Id", 42.5)

	// Print Dynamic Struct
	fmt.Println("Dynamic Struct:", dynamicType.Instance)

	// Convert Dynamic Struct → Real Struct
	type RealStruct struct {
		Name string  `json:"name" size:"3"`
		Enc  string  `json:"enc" enum:"base|ascii"`
		Id   float64 `json:"id" precision:"2"`
	}

	var realInstance RealStruct
	err = dynamicType.DynamicToRealStruct(&realInstance)
	if err != nil {
		panic(err)
	}

	// Print Real Struct
	fmt.Printf("Real Struct: %+v\n", realInstance)
}

// func TestParse(t *testing.T) {
// 	type Machine struct {
// 		// Id       float64 `json:"Id" part:"ok" precision:"2"`
// 		// Verified bool    `json:"Verified" flag:"true"`
// 		Code string `json:"Code" regex:"^[A-Z]{3}-\\d{3}$" error:"Code_must_be_in_the_format_ABC-123"`
// 	}

// 	schemaJSON, _ := StructToSchema(Machine{Code: "ABC-123"})
// 	fmt.Println("Generated Schema:", schemaJSON)

// 	dynamicType, _ := SchemaToStruct(schemaJSON)

// 	// Create an instance of the dynamic struct
// 	dynamicSchema, _ := dynamicType.ToSchema()
// 	if schemaJSON != dynamicSchema {
// 		t.Error("Schema Mismatch")
// 	} else {
// 		fmt.Println("Schema Matched")
// 	}

// 	instance := Machine{
// 		// Id: 1.23, Verified: true,
// 		Code: "ABC-123",
// 	}
// 	err := dynamicType.ValidateStruct(instance)
// 	if err != nil {
// 		t.Error("Validation Error:", err)
// 	} else {
// 		fmt.Println("Validation Passed")
// 	}

// 	data, err := json.Marshal(dynamicType.Type)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(string(data))

// 	// data, _ := json.Marshal(instance)
// 	// err = json.Unmarshal(data, &dynamicStruct.Data)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }

// 	// var helloMachine Machine
// 	// err = dynamicStruct.DynamicToRealStruct(&helloMachine)
// 	// if err != nil {
// 	// 	t.Fatal(err)
// 	// }
// 	// fmt.Printf("Real Struct: %+v\n", helloMachine)
// }
