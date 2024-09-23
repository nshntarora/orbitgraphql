package graphcache

import (
	"encoding/json"
	"regexp"
	"strings"
)

var operationNameRegex = regexp.MustCompile(`(?m)(query|mutation|subscription)\s+(\w+)\s*`)

type GraphQLRequest struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

func (gr *GraphQLRequest) Map() map[string]interface{} {
	return map[string]interface{}{
		"query":         gr.Query,
		"variables":     gr.Variables,
		"operationName": gr.OperationName,
	}
}

func (gr *GraphQLRequest) Bytes() []byte {
	bytes, _ := json.Marshal(gr)
	return bytes
}

func (gr *GraphQLRequest) FromBytes(req []byte) {
	json.Unmarshal(req, gr)
	// if the request doesn't contain an operation name, try to extract it from the query
	if gr.OperationName == "" && len(gr.Query) > 0 {
		m2 := operationNameRegex.FindString(gr.Query)
		operationNames := strings.Split(m2, " ")
		if len(operationNames) > 1 {
			gr.OperationName = strings.TrimSpace(operationNames[1])
		}
	}
}
