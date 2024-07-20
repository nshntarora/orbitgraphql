package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"graphql_cache/cache"
	"graphql_cache/transformer"
	"graphql_cache/utils"
	"graphql_cache/utils/ast_utils"
	"graphql_cache/utils/test_endpoints"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

var cacheStore = cache.NewInMemoryCache()
var recordCacheStore = cache.NewInMemoryCache()
var queryCacheStore = cache.NewInMemoryCache()

func ProxyToCachedAPI() []byte {
	copiedRequestBody := test_endpoints.REQUEST_BODY
	// Parse the query

	queryString := copiedRequestBody["query"].(string)
	astQuery, err := ast_utils.GetASTFromQuery(queryString)
	if err != nil {
		fmt.Println("Error parsing query:", err)
		return nil
	}

	cachedResponse, err := ParseASTBuildResponse(astQuery, copiedRequestBody)
	if err == nil && cachedResponse != nil {
		fmt.Println("serving response from cache...")
		br, _ := json.Marshal(cachedResponse)
		return RemoveTypenameFromResponse(br)
	}

	transformedBody, err := transformer.TransformBody(queryString, astQuery)
	if err != nil {
		fmt.Println("Error transforming body:", err)
		return nil
	}

	copiedRequestBody["query"] = transformedBody

	response := test_endpoints.GetSampleAPIResponse(copiedRequestBody)
	responseMap := make(map[string]interface{})
	err = json.Unmarshal(response, &responseMap)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
	}

	astWithTypes, err := ast_utils.GetASTFromQuery(transformedBody)
	if err != nil {
		fmt.Println("Error parsing query:", err)
		return nil
	}

	reqVariables := copiedRequestBody["variables"]
	variables := make(map[string]interface{})
	if reqVariables != nil {
		variables = reqVariables.(map[string]interface{})
	}

	for _, op := range astWithTypes.Operations {
		// for the operation op we need to traverse the response and the ast together to build a graph of the relations

		// build the relation graph
		responseKey := GetQueryResponseKey(op, responseMap, variables)
		for key, value := range responseKey {
			queryCacheStore.Set(key, value)
		}
	}

	// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
	// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
	// if the object has a nested object with __typename: "User" and id: "5678", cache
	// it as User:5678

	CacheResponse("data", responseMap, nil)

	cacheStore.Debug("cacheStore")
	recordCacheStore.Debug("recordCacheStore")
	queryCacheStore.Debug("queryCacheStore")

	// cacheState, _ := cacheStore.JSON()
	// fmt.Println(string(cacheState))

	// recordCacheState, _ := recordCacheStore.JSON()
	// fmt.Println(string(recordCacheState))

	// queryCacheState, _ := queryCacheStore.JSON()
	// fmt.Println(string(queryCacheState))

	return RemoveTypenameFromResponse(response)
}

func RemoveTypenameFromResponse(response []byte) []byte {
	mapResponse := make(map[string]interface{})
	err := json.Unmarshal(response, &mapResponse)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
		return nil
	}

	res := deleteTypename(mapResponse)

	br, _ := json.Marshal(res)
	return br
}

func deleteTypename(data interface{}) interface{} {
	switch concreteVal := data.(type) {
	case map[string]interface{}:
		// If the current item is a map, iterate through its keys.
		for key, val := range concreteVal {
			if key == "__typename" {
				// Delete the __typename key.
				delete(concreteVal, key)
			} else {
				// Recursively process nested objects/maps.
				concreteVal[key] = deleteTypename(val)
			}
		}
	case []interface{}:
		// If the current item is a slice, iterate through its elements.
		for i, val := range concreteVal {
			// Recursively process each element of the slice.
			concreteVal[i] = deleteTypename(val)
		}
	case []map[string]interface{}:
		// If the current item is a slice of maps, iterate through its elements.
		for i, val := range concreteVal {
			// Recursively process each element of the slice.
			concreteVal[i] = deleteTypename(val).(map[string]interface{})
		}
	}
	return data
}

func ParseASTBuildResponse(astQuery *ast.QueryDocument, requestBody map[string]interface{}) (interface{}, error) {

	queryDoc := astQuery.Operations[0]

	reqVariables := requestBody["variables"]
	variables := make(map[string]interface{})
	if reqVariables != nil {
		variables = reqVariables.(map[string]interface{})
	}

	queryType := queryDoc.Operation
	parentKey := queryDoc.Name

	variableDefinitions := []string{}

	for _, val := range queryDoc.VariableDefinitions {
		variableBytes, _ := json.Marshal(variables[val.Variable])
		variableDefinitions = append(variableDefinitions, val.Variable+":"+string(variableBytes))
	}

	queryResponseKey := "gql:" + string(queryType) + ":" + parentKey + "(" + hashString(strings.Join(variableDefinitions, ",")) + ")"

	cachedResponse, err := queryCacheStore.Get(queryResponseKey)
	if err == nil && cachedResponse != nil {
		switch responseType := cachedResponse.(type) {
		case string:
			res, err := TraverseResponseFromKey(responseType)
			if err != nil {
				fmt.Println("Error traversing response from key:", err)
				return nil, err
			}
			return res, nil
		case map[string]interface{}:
			finalResponse := cachedResponse.(map[string]interface{})
			for key, value := range finalResponse {
				if val, ok := value.(string); ok {
					if strings.HasPrefix(val, "gql:") {
						nestedResponse, err := TraverseResponseFromKey(val)
						if err != nil {
							fmt.Println("Error traversing nested response from key:", val, " ", err)
							return nil, err
						}
						finalResponse[key] = nestedResponse
					}
				}
			}
			return finalResponse, nil
		case []interface{}:
			responseArray := cachedResponse.([]interface{})
			for i, v := range responseArray {
				if val, ok := v.(string); ok {
					if strings.HasPrefix(val, "gql:") {
						nestedResponse, err := TraverseResponseFromKey(val)
						if err != nil {
							fmt.Println("Error traversing nested response from key:", val, " ", err)
							return nil, err
						}
						responseArray[i] = nestedResponse
					}
				} else if obj, ok := v.(map[string]interface{}); ok {
					for key, value := range obj {
						if val, ok := value.(string); ok {
							if strings.HasPrefix(val, "gql:") {
								nestedResponse, err := TraverseResponseFromKey(val)
								if err != nil {
									fmt.Println("Error traversing nested response from key:", val, " ", err)
									return nil, err
								}
								obj[key] = nestedResponse
							}
						}
					}
				}
			}
		default:
			return nil, nil
		}
	}
	return nil, nil
}

func TraverseResponseFromKey(response interface{}) (interface{}, error) {
	if val, ok := response.(string); ok {
		if strings.HasPrefix(val, "gql:") {
			response, err := cacheStore.Get(val)
			if err != nil {
				fmt.Println("Error getting response from cache:", err)
				return nil, err
			}
			return TraverseResponseFromKey(response)
		}
	}
	responseMap := response.(map[string]interface{})
	for key, value := range responseMap {
		if val, ok := value.(string); ok { // handle other data types, arrays and objects
			if strings.HasPrefix(val, "gql:") {
				nestedResponse, err := TraverseResponseFromKey(val)
				if err != nil {
					fmt.Println("Error traversing nested response from key:", val, " ", err)
					return nil, err
				}
				responseMap[key] = nestedResponse
			}
		} else if val, ok := value.(map[string]interface{}); ok {
			for k, v := range val {
				if v, ok := v.(string); ok {
					if strings.HasPrefix(v, "gql:") {
						nestedResponse, err := TraverseResponseFromKey(v)
						if err != nil {
							fmt.Println("Error traversing nested response from key:", v, " ", err)
							return nil, err
						}
						val[k] = nestedResponse
					}
				}
			}
		} else if val, ok := value.([]interface{}); ok {
			for i, v := range val {
				if v, ok := v.(string); ok {
					if strings.HasPrefix(v, "gql:") {
						nestedResponse, err := TraverseResponseFromKey(v)
						if err != nil {
							fmt.Println("Error traversing nested response from key:", v, " ", err)
							return nil, err
						}
						val[i] = nestedResponse
					}
				}
			}
		}
	}

	return responseMap, nil
}

func CacheObject(field string, object map[string]interface{}, parent map[string]interface{}) string {
	objectKeys := make([]string, 0)

	for key := range object {
		// if _, ok := val.(map[string]interface{}); !ok {
		// 	objectKeys = append(objectKeys, key)
		// }
		objectKeys = append(objectKeys, key)
	}

	parentKeys := make([]string, 0)

	if parent != nil {
		for key := range parent {
			parentKeys = append(parentKeys, key)
		}
	}

	if utils.StringArrayContainsString(objectKeys, "__typename") && utils.StringArrayContainsString(objectKeys, "id") {
		typename := object["__typename"].(string)
		id := object["id"].(string)
		cacheKey := "gql:" + typename + ":" + id
		cacheStore.Set(cacheKey, object)

		for key, value := range object {
			recordCacheStore.Set(cacheKey+":"+key, value)
		}

		return cacheKey
	} else if utils.StringArrayContainsString(objectKeys, "__typename") && !utils.StringArrayContainsString(objectKeys, "id") && parent != nil && utils.StringArrayContainsString(parentKeys, "id") && utils.StringArrayContainsString(parentKeys, "__typename") {
		typename := parent["__typename"].(string)
		parentID := parent["id"].(string)
		cacheKey := "gql:" + typename + ":" + parentID + ":" + field
		cacheStore.Set(cacheKey, object)

		for key, value := range object {
			recordCacheStore.Set(cacheKey+":"+key, value)
		}

		return cacheKey
	}

	return ""
}

func CacheResponse(field string, object map[string]interface{}, parent map[string]interface{}) (interface{}, string) {
	for key, value := range object {
		if nestedObj, ok := value.(map[string]interface{}); ok {
			_, k := CacheResponse(key, nestedObj, object)
			if k != "" {
				object[key] = k
			}
		}
		if objArray, ok := value.([]map[string]interface{}); ok {
			responseObjects := make([]interface{}, 0)
			for _, obj := range objArray {
				_, k := CacheResponse(key, obj, object)
				responseObjects = append(responseObjects, k)
			}
			if !utils.ArrayContains(responseObjects, "") {
				object[key] = responseObjects
			}
		}
		if objArray, ok := value.([]interface{}); ok {
			responseObjects := make([]interface{}, 0)
			for _, obj := range objArray {
				if objMap, ok := obj.(map[string]interface{}); ok {
					_, k := CacheResponse(key, objMap, object)
					responseObjects = append(responseObjects, k)
				}
			}
			if !utils.ArrayContains(responseObjects, "") {
				object[key] = responseObjects
			}
		}
	}

	cacheKey := CacheObject(field, object, parent)

	return object, cacheKey

	// return parent
}

func GetQueryResponseKey(queryDoc *ast.OperationDefinition, response map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	queryType := queryDoc.Operation
	parentKey := queryDoc.Name

	variableDefinitions := []string{}

	for _, val := range queryDoc.VariableDefinitions {
		variableBytes, _ := json.Marshal(variables[val.Variable])
		variableDefinitions = append(variableDefinitions, val.Variable+":"+string(variableBytes))
	}

	relationGraph := make(map[string]interface{})

	responseData := response["data"].(map[string]interface{})

	relationGraph["gql:"+string(queryType)+":"+parentKey+"("+hashString(strings.Join(variableDefinitions, ","))+")"] = GetResponseTypeID(queryDoc.SelectionSet, responseData)
	return relationGraph
}

func GetResponseTypeID(selectionSet ast.SelectionSet, response map[string]interface{}) interface{} {
	updatedSelectionSet := ast.SelectionSet{}
	// remove __typename field from the selection set
	for _, selection := range selectionSet {
		if field, ok := selection.(*ast.Field); ok {
			if field.Name != "__typename" {
				updatedSelectionSet = append(updatedSelectionSet, field)
			}
		}
	}

	if len(updatedSelectionSet) == 0 {
		return nil
	}
	if len(updatedSelectionSet) == 1 {
		// the response is an object type
		// so we will return a string with the typename and id
		// for example, Organisation:1234

		selection := updatedSelectionSet[0].(*ast.Field)
		selectionRespone := response[selection.Name].(map[string]interface{})
		if selectionRespone != nil && selectionRespone["id"] != nil && selectionRespone["__typename"] != nil {
			id := selectionRespone["id"].(string)
			typeName := selectionRespone["__typename"].(string)
			return map[string]interface{}{updatedSelectionSet[0].(*ast.Field).Name: "gql:" + typeName + ":" + id}
		}
	}
	return nil
}

func GraphSelectionSet(selectionSet ast.SelectionSet, variableDefinitions string) interface{} {

	if len(selectionSet) == 0 {
		return nil
	}

	selections := make(map[string]interface{})

	for _, selection := range selectionSet {
		switch selection := selection.(type) {
		case *ast.Field:
			selections[selection.Alias] = GraphSelectionSet(selection.SelectionSet, variableDefinitions)
		case *ast.FragmentSpread:
			// selections[selection.Name] = GraphSelectionSet(selection)
		case *ast.InlineFragment:
			// selections[selection.TypeCondition.Name] = GraphSelectionSet(selection, variableDefinitions)
		}
	}

	return selections
}

func hashVariables(variables map[string]interface{}) string {
	// base64 encode the variables
	variablesBytes, _ := json.Marshal(variables)
	return base64.StdEncoding.EncodeToString(variablesBytes)
}

func hashString(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}
