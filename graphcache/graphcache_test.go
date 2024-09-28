package graphcache

import (
	"context"
	"encoding/json"
	"testing"

	"graphql_cache/cache"

	"github.com/stretchr/testify/assert"
	"github.com/vektah/gqlparser/ast"
)

func TestNewGraphCache(t *testing.T) {
	gc := NewGraphCache()
	assert.NotNil(t, gc)
	assert.Equal(t, "id", gc.idField)
}

func TestNewGraphCacheWithOptions(t *testing.T) {
	opts := &GraphCacheOptions{
		ObjectStore: cache.NewInMemoryCache(),
		QueryStore:  cache.NewInMemoryCache(),
		Prefix:      "test::",
		IDField:     "customID",
	}
	gc := NewGraphCacheWithOptions(context.Background(), opts)
	assert.NotNil(t, gc)
	assert.Equal(t, "customID", gc.idField)
	assert.Equal(t, "test::", gc.prefix)
}

func TestKey(t *testing.T) {
	gc := NewGraphCache()
	key := gc.Key("testKey")
	assert.Equal(t, "orbit::::testKey", key)
}

func getJSONRawMessageFromMap(data map[string]interface{}) json.RawMessage {
	raw, _ := json.Marshal(data)
	return json.RawMessage(raw)
}

func TestRemoveTypenameFromResponse(t *testing.T) {
	gc := NewGraphCache()
	response := &GraphQLResponse{
		Data: getJSONRawMessageFromMap(map[string]interface{}{
			"__typename": "User",
			"id":         "123",
			"name":       "John Doe",
		}),
	}
	expectedResponse := &GraphQLResponse{
		Data: getJSONRawMessageFromMap(map[string]interface{}{
			"id":   "123",
			"name": "John Doe",
		}),
	}
	res, err := gc.RemoveTypenameFromResponse(response)
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, res)
}

func TestDeleteTypename(t *testing.T) {
	gc := NewGraphCache()
	data := map[string]interface{}{
		"__typename": "User",
		"id":         "123",
		"name":       "John Doe",
	}
	expectedData := map[string]interface{}{
		"id":   "123",
		"name": "John Doe",
	}
	res := gc.deleteTypename(data)
	assert.Equal(t, expectedData, res)
}

func TestParseASTBuildResponse(t *testing.T) {
	gc := NewGraphCache()
	astQuery := &ast.QueryDocument{
		Operations: []*ast.OperationDefinition{
			{
				Operation: ast.Query,
				Name:      "TestQuery",
			},
		},
	}
	requestBody := GraphQLRequest{
		Query:     "query TestQuery { id name }",
		Variables: map[string]interface{}{"id": "123"},
	}
	_, err := gc.ParseASTBuildResponse(astQuery, requestBody)
	assert.NotNil(t, err)
}

func TestTraverseResponseFromKey(t *testing.T) {
	gc := NewGraphCache()
	gc.cacheStore.Set(gc.Key("User:123"), map[string]interface{}{
		"id":   "123",
		"name": "John Doe",
	})
	res, err := gc.TraverseResponseFromKey(gc.Key("User:123"))
	assert.Nil(t, err)
	assert.Equal(t, map[string]interface{}{"id": "123", "name": "John Doe"}, res)
}

func TestCacheObject(t *testing.T) {
	gc := NewGraphCache()
	object := map[string]interface{}{
		"__typename": "User",
		"id":         "123",
		"name":       "John Doe",
	}
	parent := map[string]interface{}{}
	cacheKey := gc.CacheObject("user", object, parent)
	assert.Equal(t, gc.Key("User:123"), cacheKey)
}

func TestCacheResponse(t *testing.T) {
	gc := NewGraphCache()
	object := map[string]interface{}{
		"__typename": "User",
		"id":         "123",
		"name":       "John Doe",
	}
	parent := map[string]interface{}{}
	_, cacheKey := gc.CacheResponse("user", object, parent)
	assert.Equal(t, gc.Key("User:123"), cacheKey)
}

func TestCacheOperation(t *testing.T) {
	gc := NewGraphCache()
	queryDoc := &ast.OperationDefinition{
		Operation: ast.Query,
		Name:      "TestQuery",
	}
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"id":   "123",
			"name": "John Doe",
		},
	}
	variables := map[string]interface{}{"id": "123"}
	res := gc.CacheOperation(queryDoc, response, variables)
	assert.NotNil(t, res)
}

func TestGetQueryResponseKey(t *testing.T) {
	gc := NewGraphCache()
	queryDoc := &ast.OperationDefinition{
		Operation: ast.Query,
		Name:      "TestQuery",
	}
	response := map[string]interface{}{
		"data": map[string]interface{}{
			"id":   "123",
			"name": "John Doe",
		},
	}
	variables := map[string]interface{}{"id": "123"}
	res := gc.GetQueryResponseKey(queryDoc, response, variables)
	assert.NotNil(t, res)
}

func TestGetResponseTypeID(t *testing.T) {
	gc := NewGraphCache()
	selectionSet := ast.SelectionSet{
		&ast.Field{
			Name: "__typename",
		},
		&ast.Field{
			Name: "id",
		},
	}
	response := map[string]interface{}{
		"__typename": "User",
		"id":         "123",
	}
	res := gc.GetResponseTypeID(selectionSet, response)
	assert.Equal(t, map[string]interface{}{"id": gc.Key("User:123")}, res)
}

func TestGraphSelectionSet(t *testing.T) {
	gc := NewGraphCache()
	selectionSet := ast.SelectionSet{
		&ast.Field{
			Name: "id",
		},
		&ast.Field{
			Name: "name",
		},
	}
	res := gc.GraphSelectionSet(selectionSet, "")
	assert.NotNil(t, res)
}

func TestHashVariables(t *testing.T) {
	gc := NewGraphCache()
	variables := map[string]interface{}{"id": "123"}
	hash := gc.hashVariables(variables)
	assert.NotEmpty(t, hash)
}

func TestHashString(t *testing.T) {
	gc := NewGraphCache()
	str := "testString"
	hash := gc.hashString(str)
	assert.NotEmpty(t, hash)
}

func TestInvalidateCache(t *testing.T) {
	gc := NewGraphCache()
	object := map[string]interface{}{
		"__typename": "User",
		"id":         "123",
		"name":       "John Doe",
	}
	parent := map[string]interface{}{}
	_, cacheKey := gc.InvalidateCache("user", object, parent)
	assert.Equal(t, gc.Key("User:123"), cacheKey)
}

func TestInvalidateCacheObject(t *testing.T) {
	gc := NewGraphCache()
	object := map[string]interface{}{
		"__typename": "User",
		"id":         "123",
		"name":       "John Doe",
	}
	parent := map[string]interface{}{}
	cacheKey := gc.InvalidateCacheObject("user", object, parent)
	assert.Equal(t, gc.Key("User:123"), cacheKey)
}

func TestDebug(t *testing.T) {
	gc := NewGraphCache()
	assert.NotPanics(t, func() { gc.Debug() })
}

func TestLook(t *testing.T) {
	gc := NewGraphCache()
	res := gc.Look()
	assert.NotNil(t, res)
}

func TestFlush(t *testing.T) {
	gc := NewGraphCache()
	assert.NotPanics(t, func() { gc.Flush() })
}

func TestFlushByType(t *testing.T) {
	gc := NewGraphCache()
	assert.NotPanics(t, func() { gc.FlushByType("User", "123") })
}
