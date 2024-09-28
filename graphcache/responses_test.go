package graphcache

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGraphQLResponse_Map(t *testing.T) {
	data := json.RawMessage(`{"key":"value"}`)
	errors := []interface{}{"error1", "error2"}
	gr := GraphQLResponse{
		Data:   data,
		Errors: errors,
	}

	result := gr.Map()

	expected := map[string]interface{}{
		"data":   data,
		"errors": errors,
	}

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestGraphQLResponse_Bytes(t *testing.T) {
	data := json.RawMessage(`{"key":"value"}`)
	errors := []interface{}{"error1", "error2"}
	gr := GraphQLResponse{
		Data:   data,
		Errors: errors,
	}

	result := gr.Bytes()

	expected, _ := json.Marshal(gr)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", string(expected), string(result))
	}
}

func TestGraphQLResponse_FromBytes(t *testing.T) {
	jsonData := []byte(`{"data":{"key":"value"},"errors":["error1","error2"]}`)
	var gr GraphQLResponse

	gr.FromBytes(jsonData)

	expected := GraphQLResponse{
		Data:   json.RawMessage(`{"key":"value"}`),
		Errors: []interface{}{"error1", "error2"},
	}

	if !reflect.DeepEqual(gr, expected) {
		t.Errorf("expected %v, got %v", expected, gr)
	}
}
