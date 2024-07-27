package graphcache

import "encoding/json"

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []interface{}   `json:"errors"`
}

func (gr *GraphQLResponse) Map() map[string]interface{} {
	return map[string]interface{}{
		"data":   gr.Data,
		"errors": gr.Errors,
	}
}

func (gr *GraphQLResponse) Bytes() []byte {
	bytes, _ := json.Marshal(gr)
	return bytes
}

func (gr *GraphQLResponse) FromBytes(bytes []byte) {
	json.Unmarshal(bytes, gr)
}
