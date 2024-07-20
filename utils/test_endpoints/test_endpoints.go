package test_endpoints

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

var DEFAULT_PAYLOAD = strings.NewReader(`{"query":"query GetProjects($organisationId: String!, $page: Int, $perPage: Int, $query: String, $order: String, $filters: Map) {  organisation(id: $organisationId) {  id  name  projects(      page: $page      perPage: $perPage      query: $query      order: $order      filters: $filters    ) {      nodes {        id        name        description        createdAt        updatedAt        createdBy {            id            name            email        }      }      pageInfo {        page        perPage        totalPages        totalEntriesSize      }    }  }}","variables":{"organisationId":"10e89443-1f23-4ee8-8027-aaf34f904ad1","page":1,"perPage":1,"query":""}}`)

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
	fmt.Println("making api call")

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
	req.Header.Add("x-access-token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiZGFlNjUwOWQtNWVkNy00MWYzLWEzZmQtYTQzNTNiZmMzYWI5Iiwic2Vzc2lvbl9pZCI6ImI1NjcxZDJkLThiMzAtNDBhOC1iZTVkLTQyMTVjMjk0N2IxYyIsImV4cCI6MTcyNDk1NDE5Mn0.qdJ_hzJexXbN_oA4Ej5lvo-jVwvU7PqzBLfM3KpraW0")

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
