package dynamicstruct

import "testing"

func TestParse(t *testing.T) {
	jsonData := `{
		"a": { "type": "int", "enc": 3 },
		"s": { "type": "string", "size": 3 }
	}`

	Parse(jsonData)
}
