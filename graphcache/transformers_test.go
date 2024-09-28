package graphcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/ast"
)

func TestAddTypenameToQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{
			name:     "Simple query",
			query:    "{ user { id name } }",
			expected: "query { user { id name __typename } __typename }",
		},
		{
			name:     "Query with variables",
			query:    "query GetUser($id: ID!) { user(id: $id) { id name } }",
			expected: "query GetUser($id: ID!) { user(id: $id) { id name __typename } __typename }",
		},
		{
			name:     "Query with nested fields",
			query:    "{ user { id name posts { title content } } }",
			expected: "query { user { id name posts { title content __typename } __typename } __typename }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AddTypenameToQuery(tt.query)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConvertSelectionSetToString(t *testing.T) {
	selectionSet := ast.SelectionSet{
		&ast.Field{Name: "id"},
		&ast.Field{Name: "name"},
	}

	expected := "{ id name }"
	result := convertSelectionSetToString(selectionSet)
	assert.Equal(t, expected, result)
}

func TestProcessSelectionSet(t *testing.T) {
	selectionSet := ast.SelectionSet{
		&ast.Field{Name: "id"},
		&ast.Field{Name: "name"},
	}

	expected := ast.SelectionSet{
		&ast.Field{Name: "id"},
		&ast.Field{Name: "name"},
		&ast.Field{Name: "__typename"},
	}

	result := processSelectionSet(selectionSet)
	assert.Equal(t, expected, result)
}
