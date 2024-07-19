package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"graphql_cache/cache"
	"graphql_cache/utils"
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)

const NUMBER_OF_REQUESTS = 1

func main() {
	// call ProxyToCachedAPI three times and log the average time taken to get the result from the function
	// then call the ProxyToAPI three times and log the average time taken to get the result from the function
	// also compare the response bodies for both function calls, and log an error is responses are not the same

	var totalCachedTime, totalAPITime time.Duration
	var cachedResponses, apiResponses []string

	for i := 0; i < NUMBER_OF_REQUESTS; i++ {
		response, elapsed := measureExecutionTime(ProxyToAPI)
		totalAPITime += elapsed
		apiResponses = append(apiResponses, string(response))

		response, elapsed = measureExecutionTime(ProxyToCachedAPI)
		totalCachedTime += elapsed
		cachedResponses = append(cachedResponses, string(response))

	}

	avgCachedTime := totalCachedTime / NUMBER_OF_REQUESTS
	avgAPITime := totalAPITime / NUMBER_OF_REQUESTS

	fmt.Printf("Average time for ProxyToCachedAPI: %s\n", avgCachedTime)
	fmt.Printf("Average time for ProxyToAPI: %s\n", avgAPITime)

	// Compare responses
	for i := range cachedResponses {
		if cachedResponses[i] != apiResponses[i] {
			fmt.Println("Error: Response bodies are not the same")
			break
		}
	}

	fmt.Println("All responses are the same")

	fmt.Println("Writing to log files...")
	// open output file
	foc, err := os.Create("outputs-cached.log.json")
	if err != nil {
		panic(err)
	}
	// close foc on exit and check focr its returned error
	defer func() {
		if err := foc.Close(); err != nil {
			panic(err)
		}
	}()

	foc.Write([]byte("["))
	for idx, r := range cachedResponses {
		// write a chunk
		if _, err := foc.Write([]byte(r)); err != nil {
			panic(err)
		}
		if idx != len(apiResponses)-1 {
			foc.Write([]byte(","))
		}
	}
	foc.Write([]byte("]"))

	// open output file
	fod, err := os.Create("outputs-default.log.json")
	if err != nil {
		panic(err)
	}
	// close fod on exit and check fodr its returned error
	defer func() {
		if err := fod.Close(); err != nil {
			panic(err)
		}
	}()

	fod.Write([]byte("["))
	for idx, r := range apiResponses {
		// write a chunk
		if _, err := fod.Write([]byte(r)); err != nil {
			panic(err)
		}
		if idx != len(apiResponses)-1 {
			fod.Write([]byte(","))
		}
	}
	fod.Write([]byte("]"))

	fmt.Println("written to log files")
}

func measureExecutionTime(fn func() []byte) ([]byte, time.Duration) {
	start := time.Now()
	response := fn()
	elapsed := time.Since(start)
	return response, elapsed
}

var cacheStore = cache.NewInMemoryCache()
var recordCacheStore = cache.NewInMemoryCache()

var DEFAULT_PAYLOAD = strings.NewReader(`{"query":"query GetReleaseNotes(  $organisationId: String!  $page: Int  $perPage: Int  $query: String  $order: String  $filters: Map  $languageForAutoTranslation: String!) {  organisation(id: $organisationId) {    id    name    liveReleaseNotes(      page: $page      perPage: $perPage      query: $query      order: $order      filters: $filters      language: $languageForAutoTranslation    ) {      nodes {        id        title        body        alias        tags        organisationId        featuredLink        featuredLinkText        featuredImage        categories {          id          name          description          theme        }        releaseNoteId        version        pinToTop        disableComments        publishedAt  createdBy { id name }     }      pageInfo {        page        perPage        totalPages        totalEntriesSize      }    }  }}","variables":{"filters":{},"languageForAutoTranslation":"en","order":"published_at desc","organisationId":"554bb423-4ad2-4d05-b2c1-671bc249ac8c","page":1,"per_page":10,"query":"a"}}`)

func readerToMap(reader io.Reader) (map[string]interface{}, error) {
	// Read the contents of the io.Reader into a byte slice
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Unmarshal the JSON data into a map
	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

var REQUEST_BODY, _ = readerToMap(DEFAULT_PAYLOAD)

func GetSampleAPIResponse(requestBody map[string]interface{}) []byte {
	// make a graphql api call, get the response and log the time taken by the request
	url := "https://app.olvy.co/api/v2/graphql"
	method := "POST"

	// convert requestBody into a io.Reader

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Println(err)
		return nil
	}
	payload := bytes.NewReader(jsonData)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		fmt.Println(err)
		return nil
	}
	req.Header.Add("accept", "application/graphql-response+json, application/graphql+json, application/json, text/event-stream, multipart/mixed")
	req.Header.Add("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("origin", "https://releases.olvy.co")
	req.Header.Add("priority", "u=1, i")
	req.Header.Add("referer", "https://releases.olvy.co/")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return body
}

func ProxyToCachedAPI() []byte {
	copiedRequestBody := REQUEST_BODY
	transformedBody := TransformBody(copiedRequestBody)
	response := GetSampleAPIResponse(transformedBody)
	responseMap := make(map[string]interface{})
	err := json.Unmarshal(response, &responseMap)
	if err != nil {
		fmt.Println("Error unmarshalling response:", err)
	}

	// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
	// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
	// if the object has a nested object with __typename: "User" and id: "5678", cache
	// it as User:5678

	processObject(responseMap)

	cacheState, _ := cacheStore.JSON()
	fmt.Println(string(cacheState))

	recordCacheState, _ := recordCacheStore.JSON()
	fmt.Println(string(recordCacheState))

	return response
}

func cacheSingleObject(object map[string]interface{}) string {
	objectKeys := make([]string, 0)

	for key := range object {
		// if _, ok := val.(map[string]interface{}); !ok {
		// 	objectKeys = append(objectKeys, key)
		// }
		objectKeys = append(objectKeys, key)
	}

	if utils.StringArrayContainsString(objectKeys, "__typename") && utils.StringArrayContainsString(objectKeys, "id") {
		typename := object["__typename"].(string)
		id := object["id"].(string)
		cacheKey := "gql:" + typename + ":" + id
		cacheStore.Set(cacheKey, object)
		fmt.Println("update cache with:", cacheKey, " value:", object)

		for key, value := range object {
			recordCacheStore.Set(cacheKey+":"+key, value)
		}

		return cacheKey
	}

	return ""
}

func processObject(parent map[string]interface{}) (interface{}, string) {
	for key, value := range parent {
		fmt.Println("object key:", key, " value:", reflect.TypeOf(value))
		if nestedObj, ok := value.(map[string]interface{}); ok {
			_, k := processObject(nestedObj)
			parent[key] = k
		}
		if objArray, ok := value.([]map[string]interface{}); ok {
			responseObjects := make([]interface{}, 0)
			for _, obj := range objArray {
				_, k := processObject(obj)
				responseObjects = append(responseObjects, k)
			}
			parent[key] = responseObjects
		}
		if objArray, ok := value.([]interface{}); ok {
			responseObjects := make([]interface{}, 0)
			for _, obj := range objArray {
				if objMap, ok := obj.(map[string]interface{}); ok {
					_, k := processObject(objMap)
					responseObjects = append(responseObjects, k)
				}
			}
			parent[key] = responseObjects
		}
	}

	cacheKey := cacheSingleObject(parent)

	return parent, cacheKey

	// return parent
}

func ProxyToAPI() []byte {
	response := GetSampleAPIResponse(REQUEST_BODY)
	return response
}

func TransformBody(body map[string]interface{}) map[string]interface{} {
	modifiedQuery, err := AddTypenameToQuery(body["query"].(string))
	if err != nil {
		fmt.Println("Error modifying query:", err)
		return body
	}

	// transform the body to add a __typename field to every object in the query key in the graphql request
	body["query"] = modifiedQuery
	return body
}

func AddTypenameToQuery(query string) (string, error) {
	// Parse the query
	astQuery, err := parser.ParseQuery(&ast.Source{Input: query})
	if err != nil {
		fmt.Println("Error parsing query:", err)
		return "", err
	}

	// Traverse and modify the AST
	for _, operation := range astQuery.Operations {
		operation.SelectionSet = ProcessSelectionSet(operation.SelectionSet)
	}

	modifiedQuery := ""

	for _, operation := range astQuery.Operations {
		modifiedQuery += string(operation.Operation) + " "
		if operation.Name != "" {
			modifiedQuery += operation.Name + " "
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

		modifiedQuery += convertSelectionSetToString(operation.SelectionSet)
	}

	return modifiedQuery, nil
}

func convertSelectionSetToString(selectionSet ast.SelectionSet) string {
	var builder strings.Builder
	for _, selection := range selectionSet {
		switch selection := selection.(type) {
		case *ast.Field:
			builder.WriteString(selection.Name)
			builder.WriteString(" ")
			if len(selection.Arguments) > 0 {
				builder.WriteString("(")
				for i, arg := range selection.Arguments {
					if i > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(arg.Name)
					builder.WriteString(": ")
					builder.WriteString(arg.Value.String())
				}
				builder.WriteString(")")
			}
			builder.WriteString(convertSelectionSetToString(selection.SelectionSet))
		case *ast.InlineFragment:
			builder.WriteString("...")
			builder.WriteString(convertSelectionSetToString(selection.SelectionSet))
		case *ast.FragmentSpread:
			// Handle fragment spreads if necessary
		}
	}
	if builder.String() != "" {
		return "{" + builder.String() + "}"
	}
	return ""
}

func ProcessSelectionSet(selectionSet ast.SelectionSet) ast.SelectionSet {
	updatedSelectionSets := make(ast.SelectionSet, 0)
	for _, selection := range selectionSet {
		if field, ok := selection.(*ast.Field); ok {
			// Process the field
			if len(field.SelectionSet) > 0 {
				field.SelectionSet = ProcessSelectionSet(field.SelectionSet)
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
