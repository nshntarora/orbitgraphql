package graphcache

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"graphql_cache/cache"
	"graphql_cache/logger"
	"graphql_cache/utils"
	"reflect"
	"strings"

	"github.com/vektah/gqlparser/ast"
)

const DEFAULT_CACHE_PREFIX = "orbit::"
const TYPENAME_FIELD = "__typename"

// GraphCache is a struct that holds the cache stores for the GraphQL cache
type GraphCache struct {
	ctx             context.Context
	idField         string
	prefix          string
	cacheStore      cache.Cache
	queryCacheStore cache.Cache
}
type GraphCacheOptions struct {
	QueryStore  cache.Cache
	ObjectStore cache.Cache
	Prefix      string
	IDField     string
}

type CacheBackend string

const CacheBackendRedis CacheBackend = "redis"
const CacheBackendInMemory CacheBackend = "in_memory"

func NewGraphCache() *GraphCache {
	return NewGraphCacheWithOptions(context.Background(), &GraphCacheOptions{
		ObjectStore: cache.NewInMemoryCache(),
		QueryStore:  cache.NewInMemoryCache(),
	})
}

func NewGraphCacheWithOptions(ctx context.Context, opts *GraphCacheOptions) *GraphCache {
	if opts.IDField == "" {
		opts.IDField = "id"
	}
	return &GraphCache{
		ctx:             ctx,
		prefix:          opts.Prefix,
		cacheStore:      opts.ObjectStore,
		queryCacheStore: opts.QueryStore,
		idField:         opts.IDField,
	}
}

func (gc *GraphCache) Key(key string) string {
	return DEFAULT_CACHE_PREFIX + "::" + key
}

func (gc *GraphCache) RemoveTypenameFromResponse(response *GraphQLResponse) (*GraphQLResponse, error) {
	mapResponse := make(map[string]interface{})
	responseBytes, err := json.Marshal(response)
	if err != nil {
		logger.Error(gc.ctx, "Error marshalling response:", err)
		return nil, err
	}
	err = json.Unmarshal(responseBytes, &mapResponse)
	if err != nil {
		logger.Error(gc.ctx, "Error unmarshalling response: ", err, string(responseBytes))
		return nil, err
	}

	res := gc.deleteTypename(mapResponse)

	br, err := json.Marshal(res)
	if err != nil {
		logger.Error(gc.ctx, "Error marshalling response:", err)
		return nil, err
	}

	gres := GraphQLResponse{}
	err = json.Unmarshal(br, &gres)
	if err != nil {
		logger.Error(gc.ctx, "Error unmarshalling response:", err, string(br))
		return nil, err
	}

	return &gres, nil
}

func (gc *GraphCache) deleteTypename(data interface{}) interface{} {
	switch concreteVal := data.(type) {
	case map[string]interface{}:
		// If the current item is a map, iterate through its keys.
		for key, val := range concreteVal {
			if key == TYPENAME_FIELD {
				// Delete the __typename key.
				delete(concreteVal, key)
			} else {
				// Recursively process nested objects/maps.
				concreteVal[key] = gc.deleteTypename(val)
			}
		}
	case []interface{}:
		// If the current item is a slice, iterate through its elements.
		for i, val := range concreteVal {
			// Recursively process each element of the slice.
			concreteVal[i] = gc.deleteTypename(val)
		}
	case []map[string]interface{}:
		// If the current item is a slice of maps, iterate through its elements.
		for i, val := range concreteVal {
			// Recursively process each element of the slice.
			concreteVal[i] = gc.deleteTypename(val).(map[string]interface{})
		}
	}
	return data
}

func (gc *GraphCache) ParseASTBuildResponse(astQuery *ast.QueryDocument, requestBody GraphQLRequest) (interface{}, error) {

	if len(astQuery.Operations) == 0 {
		return nil, errors.New("no operations found in query")
	}

	queryDoc := astQuery.Operations[0]

	reqVariables := requestBody.Variables
	variables := make(map[string]interface{})
	if reqVariables != nil {
		variables = reqVariables
	}

	queryType := queryDoc.Operation
	parentKey := queryDoc.Name

	variableDefinitions := []string{}

	for _, val := range queryDoc.VariableDefinitions {
		variableBytes, _ := json.Marshal(variables[val.Variable])
		variableDefinitions = append(variableDefinitions, val.Variable+":"+string(variableBytes))
	}

	queryResponseKey := gc.Key(string(queryType) + ":" + parentKey + "(" + gc.hashString(strings.Join(variableDefinitions, ",")) + ")" + gc.prefix)

	cachedResponse, err := gc.queryCacheStore.Get(queryResponseKey)
	if err == nil && cachedResponse != nil {
		switch responseType := cachedResponse.(type) {
		case string:
			res, err := gc.TraverseResponseFromKey(responseType)
			if err != nil || res == nil {
				logger.Error(gc.ctx, "Error traversing response from key:", err)
				return nil, err
			}
			return res, nil
		case map[string]interface{}:
			finalResponse := cachedResponse.(map[string]interface{})
			for key, value := range finalResponse {
				if val, ok := value.(string); ok {
					if strings.HasPrefix(val, DEFAULT_CACHE_PREFIX) {
						nestedResponse, err := gc.TraverseResponseFromKey(val)
						if err != nil || nestedResponse == nil {
							logger.Error(gc.ctx, "Error traversing nested response from key:", val, " ", err)
							return nil, err
						}
						finalResponse[key] = nestedResponse
					}
				}
				if val, ok := value.(map[string]interface{}); ok {
					for k, v := range val {
						if v, ok := v.(string); ok {
							if strings.HasPrefix(v, DEFAULT_CACHE_PREFIX) {
								nestedResponse, err := gc.TraverseResponseFromKey(v)
								if err != nil || nestedResponse == nil {
									logger.Error(gc.ctx, "Error traversing nested response from key:", v, " ", err)
									return nil, err
								}
								val[k] = nestedResponse
							}
						}
					}
					finalResponse[key] = val
				}
				if val, ok := value.([]interface{}); ok {
					for i, v := range val {
						if v, ok := v.(string); ok {
							if strings.HasPrefix(v, DEFAULT_CACHE_PREFIX) {
								nestedResponse, err := gc.TraverseResponseFromKey(v)
								if err != nil || nestedResponse == nil {
									logger.Error(gc.ctx, "Error traversing nested response from key:", v, " ", err)
									return nil, err
								}
								val[i] = nestedResponse
							}
						}
					}
					finalResponse[key] = val

				}
				if val, ok := value.([]map[string]interface{}); ok {
					for i, v := range val {
						for k, v := range v {
							if v, ok := v.(string); ok {
								if strings.HasPrefix(v, DEFAULT_CACHE_PREFIX) {
									nestedResponse, err := gc.TraverseResponseFromKey(v)
									if err != nil || nestedResponse == nil {
										logger.Error(gc.ctx, "Error traversing nested response from key:", v, " ", err)
										return nil, err
									}
									val[i][k] = nestedResponse
								}
							}
						}
					}
					finalResponse[key] = val
				}
			}
			return finalResponse, nil
		case []interface{}:
			responseArray := cachedResponse.([]interface{})
			for i, v := range responseArray {
				if val, ok := v.(string); ok {
					if strings.HasPrefix(val, DEFAULT_CACHE_PREFIX) {
						nestedResponse, err := gc.TraverseResponseFromKey(val)
						if err != nil || nestedResponse == nil {
							logger.Error(gc.ctx, "Error traversing nested response from key:", val, " ", err)
							return nil, err
						}
						responseArray[i] = nestedResponse
					}
				} else if obj, ok := v.(map[string]interface{}); ok {
					for key, value := range obj {
						if val, ok := value.(string); ok {
							if strings.HasPrefix(val, DEFAULT_CACHE_PREFIX) {
								nestedResponse, err := gc.TraverseResponseFromKey(val)
								if err != nil || nestedResponse == nil {
									logger.Error(gc.ctx, "Error traversing nested response from key:", val, " ", err)
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
	return nil, errors.New("error getting response from cache")
}

func (gc *GraphCache) TraverseResponseFromKey(response interface{}) (interface{}, error) {
	if val, ok := response.(string); ok {
		if strings.HasPrefix(val, DEFAULT_CACHE_PREFIX) {
			response, err := gc.cacheStore.Get(val)
			if err != nil {
				logger.Error(gc.ctx, "Error getting response from cache:", err)
				return nil, err
			}
			return gc.TraverseResponseFromKey(response)
		}
	} else if responseMap, ok := response.(map[string]interface{}); ok {
		for key, value := range responseMap {
			if val, ok := value.(string); ok { // handle other data types, arrays and objects
				if strings.HasPrefix(val, DEFAULT_CACHE_PREFIX) {
					nestedResponse, err := gc.TraverseResponseFromKey(val)
					if err != nil {
						logger.Error(gc.ctx, "Error traversing nested response from key:", val, " ", err)
						return nil, err
					}
					responseMap[key] = nestedResponse
				}
			} else if val, ok := value.(map[string]interface{}); ok {
				for k, v := range val {
					if v, ok := v.(string); ok {
						if strings.HasPrefix(v, DEFAULT_CACHE_PREFIX) {
							nestedResponse, err := gc.TraverseResponseFromKey(v)
							if err != nil {
								logger.Error(gc.ctx, "Error traversing nested response from key:", v, " ", err)
								return nil, err
							}
							val[k] = nestedResponse
						}
					}
				}
			} else if val, ok := value.([]interface{}); ok {
				for i, v := range val {
					if v, ok := v.(string); ok {
						if strings.HasPrefix(v, DEFAULT_CACHE_PREFIX) {
							nestedResponse, err := gc.TraverseResponseFromKey(v)
							if err != nil {
								logger.Error(gc.ctx, "Error traversing nested response from key:", v, " ", err)
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

	return nil, errors.New("error traversing response from key")
}

func (gc *GraphCache) CacheObject(field string, object map[string]interface{}, parent map[string]interface{}) string {
	objectKeys := make([]string, 0)

	for key := range object {
		// if _, ok := val.(map[string]interface{}); !ok {
		// 	objectKeys = append(objectKeys, key)
		// }
		objectKeys = append(objectKeys, key)
	}

	parentKeys := make([]string, 0)

	for key := range parent {
		parentKeys = append(parentKeys, key)
	}

	if utils.StringArrayContainsString(objectKeys, TYPENAME_FIELD) && utils.StringArrayContainsString(objectKeys, gc.idField) {
		typename := object[TYPENAME_FIELD].(string)
		id := object[gc.idField].(string)
		cacheKey := typename + ":" + id
		gc.cacheStore.Set(gc.Key(cacheKey), object)
		return gc.Key(cacheKey)
	} else if utils.StringArrayContainsString(objectKeys, TYPENAME_FIELD) && !utils.StringArrayContainsString(objectKeys, gc.idField) && parent != nil && utils.StringArrayContainsString(parentKeys, gc.idField) && utils.StringArrayContainsString(parentKeys, TYPENAME_FIELD) {
		typename := parent[TYPENAME_FIELD].(string)
		parentID := parent[gc.idField].(string)
		cacheKey := typename + ":" + parentID + ":" + field
		gc.cacheStore.Set(gc.Key(cacheKey), object)
		return gc.Key(cacheKey)
	}

	return ""
}

func (gc *GraphCache) CacheResponse(field string, object map[string]interface{}, parent map[string]interface{}) (interface{}, string) {
	for key, value := range object {
		if nestedObj, ok := value.(map[string]interface{}); ok {
			_, k := gc.CacheResponse(key, nestedObj, object)
			if k != "" {
				object[key] = k
			}
		}
		if objArray, ok := value.([]map[string]interface{}); ok {
			responseObjects := make([]interface{}, 0)
			for _, obj := range objArray {
				_, k := gc.CacheResponse(key, obj, object)
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
					_, k := gc.CacheResponse(key, objMap, object)
					responseObjects = append(responseObjects, k)
				} else {
					appendToInterfaceArray(obj, &responseObjects)
				}
			}
			if !utils.ArrayContains(responseObjects, "") {
				object[key] = responseObjects
			}
		}
	}

	cacheKey := gc.CacheObject(field, object, parent)

	return object, cacheKey
}

func appendToInterfaceArray[T any](obj interface{}, responseObjects *[]T) {
	switch v := obj.(type) {
	case T:
		*responseObjects = append(*responseObjects, v)
	}
}

func (gc *GraphCache) CacheOperation(queryDoc *ast.OperationDefinition, response map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	responseKey := gc.GetQueryResponseKey(queryDoc, response, variables)
	for key, value := range responseKey {
		if value != nil {
			gc.queryCacheStore.Set(key, value)
		}
	}
	return responseKey
}

func (gc *GraphCache) GetQueryResponseKey(queryDoc *ast.OperationDefinition, response map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	queryType := queryDoc.Operation
	parentKey := queryDoc.Name

	variableDefinitions := []string{}

	for _, val := range queryDoc.VariableDefinitions {
		variableBytes, _ := json.Marshal(variables[val.Variable])
		variableDefinitions = append(variableDefinitions, val.Variable+":"+string(variableBytes))
	}

	relationGraph := make(map[string]interface{})

	if response == nil || response["data"] == nil {
		return relationGraph
	}

	responseData := response["data"].(map[string]interface{})

	relationGraph[gc.Key(string(queryType)+":"+parentKey+"("+gc.hashString(strings.Join(variableDefinitions, ","))+")"+gc.prefix)] = gc.GetResponseTypeID(queryDoc.SelectionSet, responseData)
	return relationGraph
}

func (gc *GraphCache) GetResponseTypeID(selectionSet ast.SelectionSet, response map[string]interface{}) interface{} {
	updatedSelectionSet := ast.SelectionSet{}
	// remove __typename field from the selection set
	for _, selection := range selectionSet {
		if field, ok := selection.(*ast.Field); ok {
			if field.Name != TYPENAME_FIELD {
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

		if response[selection.Name] == nil {
			return nil
		}

		switch reflect.TypeOf(response[selection.Name]).Kind() {
		case reflect.Map:
			selectionRespone, ok := response[selection.Name].(map[string]interface{})
			if ok && selectionRespone != nil && selectionRespone[gc.idField] != nil && selectionRespone[TYPENAME_FIELD] != nil {
				id := selectionRespone[gc.idField].(string)
				typeName := selectionRespone[TYPENAME_FIELD].(string)
				return map[string]interface{}{updatedSelectionSet[0].(*ast.Field).Name: gc.Key(typeName + ":" + id)}
			}
		case reflect.Slice:
			selectionRespone, ok := response[selection.Name].([]interface{})
			if ok && selectionRespone != nil {
				responseObjects := make([]interface{}, 0)
				for _, obj := range selectionRespone {
					if objMap, ok := obj.(map[string]interface{}); ok {
						if objMap[gc.idField] != nil && objMap[TYPENAME_FIELD] != nil {
							id := objMap[gc.idField].(string)
							typeName := objMap[TYPENAME_FIELD].(string)
							responseObjects = append(responseObjects, gc.Key(typeName+":"+id))
						}
					}
				}
				return map[string]interface{}{updatedSelectionSet[0].(*ast.Field).Name: responseObjects}
			}
		}
	}
	return nil
}

func (gc *GraphCache) GraphSelectionSet(selectionSet ast.SelectionSet, variableDefinitions string) interface{} {

	if len(selectionSet) == 0 {
		return nil
	}

	selections := make(map[string]interface{})

	for _, selection := range selectionSet {
		switch selection := selection.(type) {
		case *ast.Field:
			selections[selection.Alias] = gc.GraphSelectionSet(selection.SelectionSet, variableDefinitions)
		case *ast.FragmentSpread:
			// selections[selection.Name] = GraphSelectionSet(selection)
		case *ast.InlineFragment:
			// selections[selection.TypeCondition.Name] = GraphSelectionSet(selection, variableDefinitions)
		}
	}

	return selections
}

func (gc *GraphCache) hashVariables(variables map[string]interface{}) string {
	// base64 encode the variables
	variablesBytes, _ := json.Marshal(variables)
	return base64.StdEncoding.EncodeToString(variablesBytes)
}

func (gc *GraphCache) hashString(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (gc *GraphCache) InvalidateCache(field string, object map[string]interface{}, parent map[string]interface{}) (interface{}, string) {
	for key, value := range object {
		if nestedObj, ok := value.(map[string]interface{}); ok {
			_, k := gc.InvalidateCache(key, nestedObj, object)
			if k != "" {
				object[key] = k
			}
		}
		if objArray, ok := value.([]map[string]interface{}); ok {
			responseObjects := make([]interface{}, 0)
			for _, obj := range objArray {
				_, k := gc.InvalidateCache(key, obj, object)
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
					_, k := gc.InvalidateCache(key, objMap, object)
					responseObjects = append(responseObjects, k)
				}
			}
			if !utils.ArrayContains(responseObjects, "") {
				object[key] = responseObjects
			}
		}
	}

	cacheKey := gc.InvalidateCacheObject(field, object, parent)

	return object, cacheKey

	// return parent
}

func (gc *GraphCache) InvalidateCacheObject(field string, object map[string]interface{}, parent map[string]interface{}) string {
	objectKeys := make([]string, 0)

	for key := range object {
		// if _, ok := val.(map[string]interface{}); !ok {
		// 	objectKeys = append(objectKeys, key)
		// }
		objectKeys = append(objectKeys, key)
	}

	parentKeys := make([]string, 0)

	for key := range parent {
		parentKeys = append(parentKeys, key)
	}

	if utils.StringArrayContainsString(objectKeys, TYPENAME_FIELD) && utils.StringArrayContainsString(objectKeys, gc.idField) {
		typename := object[TYPENAME_FIELD].(string)
		id := object[gc.idField].(string)
		cacheKey := typename + ":" + id
		gc.cacheStore.DeleteByPrefix(gc.Key(cacheKey))
		return gc.Key(cacheKey)
	} else if utils.StringArrayContainsString(objectKeys, TYPENAME_FIELD) && !utils.StringArrayContainsString(objectKeys, gc.idField) && parent != nil && utils.StringArrayContainsString(parentKeys, gc.idField) && utils.StringArrayContainsString(parentKeys, TYPENAME_FIELD) {
		typename := parent[TYPENAME_FIELD].(string)
		parentID := parent[gc.idField].(string)
		cacheKey := typename + ":" + parentID + ":" + field
		gc.cacheStore.DeleteByPrefix(gc.Key(cacheKey))
		return gc.Key(cacheKey)
	}

	return ""
}

func (gc *GraphCache) Debug() {
	gc.cacheStore.Debug("cacheStore")
	gc.queryCacheStore.Debug("queryCacheStore")
}

func (gc *GraphCache) Look() map[string]interface{} {
	output := make(map[string]interface{})
	cacheMap, _ := gc.cacheStore.Map()
	queryCacheMap, _ := gc.queryCacheStore.Map()

	output["cacheStore"] = cacheMap
	output["queryCacheStore"] = queryCacheMap

	return output
}

func (gc *GraphCache) Flush() {
	gc.cacheStore.Flush()
	gc.queryCacheStore.Flush()
}

func (gc *GraphCache) FlushByType(typeName string, id string) {
	gc.cacheStore.DeleteByPrefix(gc.Key(typeName + ":" + id))
	gc.queryCacheStore.DeleteByPrefix(gc.Key(typeName + ":" + id))
}
