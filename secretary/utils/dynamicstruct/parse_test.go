package dynamicstruct

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/codeharik/secretary/utils/binstruct"
)

type Machine struct {
	Id       float64 `json:"Id" part:"ok" precision:"2"`
	Verified bool    `json:"Verified" flag:"true"`
	Fruit    string  `json:"Fruit" bin:"Fruit" regex:"^(apple|banana|cherry)$"`
	Code     string  `json:"Code" bin:"Code" regex:"^[A-Z]{3}-\\d{3}$" error:"Code_must_be_in_the_format_ABC-123"`
}

func TestParse(t *testing.T) {
	tests := []struct {
		pass    bool
		machine Machine
	}{
		{
			true,
			Machine{Id: 1.23, Verified: true, Code: "ABC-123", Fruit: "cherry"},
		},
		// {
		// 	false,
		// 	Machine{Id: 1.23, Verified: true, Code: "ABsC-123", Fruit: "cherry"},
		// },
		// {
		// 	false,
		// 	Machine{Id: 1.23, Verified: true, Fruit: "cherry"},
		// },
		// {
		// 	false,
		// 	Machine{Id: 1.23, Verified: true, Code: "ABC-123"},
		// },
		// {
		// 	false,
		// 	Machine{Id: 1.23, Verified: true},
		// },
	}

	for _, test := range tests {

		machineSchema, _ := StructToSchema(test.machine)

		dynamicMachine, _ := SchemaToStruct(machineSchema)

		// Create an instance of the dynamic struct
		dynamicSchema, _ := dynamicMachine.ToSchema()
		if machineSchema != dynamicSchema {
			t.Error("Schema Mismatch")
		}

		testMachine(t, dynamicMachine, test.pass, test.machine)

		dynamicMachine, _ = SchemaToStruct(machineSchema)
		dynamicMachine.SetField("Fruit", "apple")
		dynamicMachine.SetField("Code", "ABC-123")
		var newMachine Machine
		err := dynamicMachine.DynamicToRealStruct(&newMachine)
		if err != nil {
			t.Fatal(err)
		}
		dynamicMachine, _ = SchemaToStruct(machineSchema)
		testMachine(t, dynamicMachine, true, newMachine)
	}
}

func testMachine(t *testing.T, dynamicMachine *DynamicStruct, pass bool, machine Machine) {
	err := dynamicMachine.ValidateStruct(machine)
	if pass && (err != nil) {
		t.Error("Validation Error:", err)
	}

	err = dynamicMachine.Validate()
	if err == nil {
		t.Error("Validation Error:", err)
	}

	machineJson, err := json.Marshal(machine)
	if err != nil {
		t.Error(err)
	}
	err = dynamicMachine.JsonUnmarshal(machineJson)
	if err != nil {
		t.Error(err)
	}
	err = dynamicMachine.Validate()
	if pass && (err != nil) {
		t.Error("Validation Error:", err)
	}

	var newMachine Machine
	err = dynamicMachine.DynamicToRealStruct(&newMachine)
	if err != nil {
		t.Fatal(err)
	}

	err = dynamicMachine.ValidateStruct(newMachine)
	if pass && (err != nil) {
		t.Error("Validation Error:", err)
	}

	if pass && (err != nil || reflect.DeepEqual(machine, newMachine) == false) {
		t.Errorf("Json Mismatch %v", err)
	}

	m, _ := binstruct.Serialize(machine)
	n, _ := binstruct.Serialize(newMachine)
	d, _ := binstruct.Serialize(dynamicMachine.Instance.Interface())
	if string(m) != string(n) || string(m) != string(d) {
		t.Error("Binary Mismatch")
	}
	// var newMachineM Machine
	// if err = binstruct.Deserialize(m, &newMachineM); err != nil {
	// 	t.Fatal(err)
	// }
	// var newMachineN Machine
	// if err = binstruct.Deserialize(n, &newMachineN); err != nil {
	// 	t.Fatal(err)
	// }
	// var newMachineD Machine
	// if err = binstruct.Deserialize(d, &newMachineD); err != nil {
	// 	t.Fatal(err)
	// }
	// utils.Log("machine", machine,
	// 	"m", string(m),
	// 	"n", string(n),
	// 	"d", string(d),
	// 	"m", m,
	// 	"n", n,
	// 	"d", d,
	// 	"newMachineM", newMachineM,
	// 	"newMachineN", newMachineN,
	// 	"newMachineD", newMachineD)
}
