package graphcache

import (
	"github.com/vektah/gqlparser/ast"
	"github.com/vektah/gqlparser/parser"
)

func GetASTFromQuery(query string) (*ast.QueryDocument, error) {
	doc, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		return nil, err
	}
	return doc, nil
}
