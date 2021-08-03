package gojsonschema

import (
	"fmt"
	"testing"
)

const testingSchema = `{
    "$schema": "http://json-schema.org/draft-04/schema#",
    "type": "object",
    "oneOf": [
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {"A": {"type": "number"}},
            "required": ["A"]
        },
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {"B": {"type": "number"}},
            "required": ["B"]
        }
    ],
    "anyOf": [
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {"C": {"type": "number"}},
            "required": ["C"]
        },
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {"D": {"type": "number"}},
            "required": ["D"]
        },
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {
                "G": {
                    "type": "object",
                    "additionalProperties": false,
                    "properties": {"A": {"type": "number"}},
                    "required": ["A"]
                }
            },
            "required": ["G"]
        }
    ],
    "allOf": [
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {"E": {"type": "number"}},
            "required": ["E"]
        },
        {
            "$schema": "http://json-schema.org/draft-04/schema#",
            "type": "object",
            "additionalProperties": false,
            "properties": {"F": {"type": "number"}},
            "required": ["F"]
        }
    ]
}`

func TestAdditionalPropertiesFalse(t *testing.T) {
	schemaLoader := NewStringLoader(testingSchema)
	s, err := NewSchema(schemaLoader)
	if err != nil {
		t.Fatalf("error compiling schema: %v", err)
	}

	var tests = []struct {
		description string
		document    string
		valid       bool
		err         error
		score       int
		errors      []string
	}{
		{
			description: "valid document",
			document:    `{"A":1,"C":3,"D":4,"E":5,"F":6,"G":{"A":71}}`,
			valid:       true,
			err:         nil,
			score:       22,
			errors:      nil,
		},
		{
			description: "invalid document; check for field overlap",
			document:    `{"A":1,"C":3,"D":4,"E":5,"F":6,"G":{"C":73}}`,
			valid:       false,
			err:         nil,
			score:       21,
			errors: []string{
				"(root): Must validate one and only one schema (oneOf)",
				"(root): Additional property G is not allowed",
				"(root): Additional property G is not allowed",
				"(root): Additional property G is not allowed",
				"(root): Must validate all the schemas (allOf)",
			},
		},
		{
			description: "invalid document; nothing present",
			document:    `{}`,
			valid:       false,
			err:         nil,
			score:       6,
			errors: []string{
				"(root): Must validate at least one schema (anyOf)",
				"(root): C is required",
				"(root): Must validate one and only one schema (oneOf)",
				"(root): A is required",
				"(root): E is required",
				"(root): F is required",
				"(root): Must validate all the schemas (allOf)",
			},
		},
		{
			description: "invalid document; unspecified field 'Z'",
			document:    `{"A":1,"C":3,"D":4,"E":5,"F":6,"Z":7}`,
			valid:       false,
			err:         nil,
			score:       20,
			errors: []string{
				"(root): Must validate at least one schema (anyOf)",
				"(root): Additional property Z is not allowed",
				"(root): Must validate one and only one schema (oneOf)",
				"(root): Additional property D is not allowed",
				"(root): Additional property Z is not allowed",
				"(root): Additional property D is not allowed",
				"(root): Additional property Z is not allowed",
				"(root): Additional property D is not allowed",
				"(root): Additional property Z is not allowed",
				"(root): Must validate all the schemas (allOf)",
			},
		},
		{
			description: "invalid document; both 'oneOf' present",
			document:    `{"A":1,"B":2,"C":3,"D":4,"E":5,"F":6}`,
			valid:       false,
			err:         nil,
			score:       21,
			errors: []string{
				"(root): Must validate one and only one schema (oneOf)",
				"(root): Additional property B is not allowed",
				"(root): Additional property B is not allowed",
				"(root): Additional property B is not allowed",
				"(root): Must validate all the schemas (allOf)",
			},
		},
		{
			description: "invalid document; neither 'oneOf' present",
			document:    `{"C":3,"D":4,"E":5,"F":6}`,
			valid:       false,
			err:         nil,
			score:       22,
			errors: []string{
				"(root): Must validate one and only one schema (oneOf)",
				"(root): A is required",
			},
		},
		{
			description: "invalid document; neither 'anyOf' present",
			document:    `{"A":1,"E":5,"F":6}`,
			valid:       false,
			err:         nil,
			score:       22,
			errors: []string{
				"(root): Must validate at least one schema (anyOf)",
				"(root): C is required",
			},
		},
		{
			description: "invalid document; one 'allOf' not present",
			document:    `{"A":1,"C":3,"D":4,"E":5}`,
			valid:       false,
			err:         nil,
			score:       13,
			errors: []string{
				"(root): F is required",
				"(root): Must validate all the schemas (allOf)",
			},
		},
		{
			description: "invalid document; neither 'allOf' present",
			document:    `{"A":1,"C":3,"D":4}`,
			valid:       false,
			err:         nil,
			score:       6,
			errors: []string{
				"(root): E is required",
				"(root): F is required",
				"(root): Must validate all the schemas (allOf)",
			},
		},
	}
	for _, test := range tests {
		documentLoader := NewStringLoader(test.document)
		result, err := s.Validate(documentLoader)
		fmt.Printf("%q: valid:%t err:%v result:%v\n", test.description, result.Valid(), err, result)

		if (test.err != nil) != (err != nil) {
			t.Errorf("test %q unexpected error\nwant: %v\ngot : %v", test.description, test.err, err)
		}
		if test.valid != result.Valid() {
			t.Errorf("test %q unexpected validity\nwant: %v\ngot : %v", test.description, test.valid, result.Valid())
		}
		if test.score != result.score {
			t.Errorf("test %q unexpected score\nwant: %d\ngot : %d", test.description, test.score, result.score)
		}

		var gotErrors, wantErrors string
		for _, err := range test.errors {
			wantErrors += err + "\n"
		}
		for _, err := range result.Errors() {
			gotErrors += err.String() + "\n"
		}
		if gotErrors != wantErrors {
			t.Errorf("test %q unexpected errors\nwant: %s\ngot : %s", test.description, wantErrors, gotErrors)
		}
	}
}
