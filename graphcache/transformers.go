package graphcache

import (
	"strings"

	"github.com/vektah/gqlparser/ast"
)

func AddTypenameToQuery(query string) (string, error) {

	astQuery, err := GetASTFromQuery(query)
	if err != nil {
		return "", err
	}

	// Traverse and modify the AST
	for _, operation := range astQuery.Operations {
		operation.SelectionSet = processSelectionSet(operation.SelectionSet)
	}

	modifiedQuery := ""

	for _, operation := range astQuery.Operations {
		modifiedQuery += string(operation.Operation) + " "
		if operation.Name != "" {
			modifiedQuery += operation.Name
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
		// modifiedQuery += " "

		modifiedQuery += convertSelectionSetToString(operation.SelectionSet)
	}

	return modifiedQuery, nil
}

func convertSelectionSetToString(selectionSet ast.SelectionSet) string {
	var builder []string
	for _, selection := range selectionSet {
		switch selection := selection.(type) {
		case *ast.Field:
			field := []string{}
			field = append(field, selection.Name)
			if len(selection.Arguments) > 0 {
				args := []string{}
				args = append(args, "(")
				for i, arg := range selection.Arguments {
					if i > 0 {
						args = append(args, ", ")
					}
					args = append(args, arg.Name)
					args = append(args, ": ")
					args = append(args, arg.Value.String())
				}
				args = append(args, ")")
				field = append(field, strings.Join(args, ""))
			}
			builder = append(builder, strings.Join(field, ""))
			if len(selection.SelectionSet) > 0 {
				builder = append(builder, convertSelectionSetToString(selection.SelectionSet))
			}
		case *ast.InlineFragment:
			builder = append(builder, "...")
			if len(selection.SelectionSet) > 0 {
				builder = append(builder, convertSelectionSetToString(selection.SelectionSet))
			}
		case *ast.FragmentSpread:
			// Handle fragment spreads if necessary
		}
	}
	if len(builder) > 0 {
		block := []string{}
		block = append(block, "{")
		block = append(block, builder...)
		block = append(block, "}")
		return strings.Join(block, " ")
	}
	return ""
}

func processSelectionSet(selectionSet ast.SelectionSet) ast.SelectionSet {
	updatedSelectionSets := make(ast.SelectionSet, 0)
	for _, selection := range selectionSet {
		if field, ok := selection.(*ast.Field); ok {
			// Process the field
			if len(field.SelectionSet) > 0 {
				field.SelectionSet = processSelectionSet(field.SelectionSet)
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
