package handlers

import (
	"graphql_cache/utils/test_endpoints"
)

func ProxyToAPI() []byte {
	response := test_endpoints.GetSampleAPIResponse(test_endpoints.REQUEST_BODY)
	return response
}
