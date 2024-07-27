package graphcache

import "encoding/json"

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

func (gr *GraphQLRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"query":     gr.Query,
		"variables": gr.Variables,
	}
}

func (gr *GraphQLRequest) Bytes() []byte {
	bytes, _ := json.Marshal(gr)
	return bytes
}
