package transformer

import (
	"fmt"
	"graphql_cache/utils/ast_utils"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

func TransformBody(queryString string, ast *ast.QueryDocument) (string, error) {
	modifiedQuery, err := AddTypenameToQuery(queryString)
	if err != nil {
		fmt.Println("Error modifying query:", err)
		return "", err
	}

	return modifiedQuery, nil
}

func AddTypenameToQuery(query string) (string, error) {

	astQuery, err := ast_utils.GetASTFromQuery(query)
	if err != nil {
		return "", err
	}

	// Traverse and modify the AST
	for _, operation := range astQuery.Operations {
		operation.SelectionSet = ProcessSelectionSet(operation.SelectionSet)
	}

	modifiedQuery := ""

	for _, operation := range astQuery.Operations {
		modifiedQuery += string(operation.Operation) + " "
		if operation.Name != "" {
			modifiedQuery += operation.Name + " "
		}
		if len(operation.VariableDefinitions) > 0 {
			modifiedQuery += "("
			for i, variable := range operation.VariableDefinitions {
				if i > 0 {
					modifiedQuery += ", "
				}
				modifiedQuery += "$" + variable.Variable + ": " + variable.Type.Name()
				if variable.Type.NonNull {
					modifiedQuery += "!"
				}
			}
			modifiedQuery += ") "
		}

		modifiedQuery += convertSelectionSetToString(operation.SelectionSet)
	}

	return modifiedQuery, nil
}

func convertSelectionSetToString(selectionSet ast.SelectionSet) string {
	var builder strings.Builder
	for _, selection := range selectionSet {
		switch selection := selection.(type) {
		case *ast.Field:
			builder.WriteString(selection.Name)
			builder.WriteString(" ")
			if len(selection.Arguments) > 0 {
				builder.WriteString("(")
				for i, arg := range selection.Arguments {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(arg.Name)
					builder.WriteString(": ")
					builder.WriteString(arg.Value.String())
				}
				builder.WriteString(")")
			}
			builder.WriteString(convertSelectionSetToString(selection.SelectionSet))
		case *ast.InlineFragment:
			builder.WriteString("...")
			builder.WriteString(convertSelectionSetToString(selection.SelectionSet))
		case *ast.FragmentSpread:
			// Handle fragment spreads if necessary
		}
	}
	if builder.String() != "" {
		return "{" + builder.String() + "}"
	}
	return ""
}

func ProcessSelectionSet(selectionSet ast.SelectionSet) ast.SelectionSet {
	updatedSelectionSets := make(ast.SelectionSet, 0)
	for _, selection := range selectionSet {
		if field, ok := selection.(*ast.Field); ok {
			// Process the field
			if len(field.SelectionSet) > 0 {
				field.SelectionSet = ProcessSelectionSet(field.SelectionSet)
			}
			updatedSelectionSets = append(updatedSelectionSets, field)
		}
	}

	if len(updatedSelectionSets) > 0 {
		exists := false
		for _, s := range updatedSelectionSets {
			if field, ok := s.(*ast.Field); ok && field.Name == "__typename" {
				exists = true
				break
			}
		}
		if !exists {
			// Add __typename
			typenameField := &ast.Field{
				Name: "__typename",
			}
			updatedSelectionSets = append(updatedSelectionSets, typenameField)
		}
	}
	return updatedSelectionSets
}
