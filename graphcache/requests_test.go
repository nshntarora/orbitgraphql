package graphcache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphQLRequest_FromBytes_WithOperationName(t *testing.T) {
	req := []byte(`{"operationName":"TestQuery","query":"query TestQuery { id name }","variables":{"id":"123"}}`)
	var gqlReq GraphQLRequest
	gqlReq.FromBytes(req)

	assert.Equal(t, "TestQuery", gqlReq.OperationName)
	assert.Equal(t, "query TestQuery { id name }", gqlReq.Query)
	assert.Equal(t, map[string]interface{}{"id": "123"}, gqlReq.Variables)
}

func TestGraphQLRequest_FromBytes_WithoutOperationName(t *testing.T) {
	req := []byte(`{"query":"query TestQuery { id name }","variables":{"id":"123"}}`)
	var gqlReq GraphQLRequest
	gqlReq.FromBytes(req)

	assert.Equal(t, "TestQuery", gqlReq.OperationName)
	assert.Equal(t, "query TestQuery { id name }", gqlReq.Query)
	assert.Equal(t, map[string]interface{}{"id": "123"}, gqlReq.Variables)
}

func TestGraphQLRequest_FromBytes_InvalidJSON(t *testing.T) {
	req := []byte(`{"query":"query TestQuery { id name }","variables":{"id":"123"}`)
	var gqlReq GraphQLRequest
	gqlReq.FromBytes(req)

	assert.Empty(t, gqlReq.OperationName)
	assert.Empty(t, gqlReq.Query)
	assert.Nil(t, gqlReq.Variables)
}

func TestGraphQLRequest_FromBytes_EmptyJSON(t *testing.T) {
	req := []byte(`{}`)
	var gqlReq GraphQLRequest
	gqlReq.FromBytes(req)

	assert.Empty(t, gqlReq.OperationName)
	assert.Empty(t, gqlReq.Query)
	assert.Nil(t, gqlReq.Variables)
}
