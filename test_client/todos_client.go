package test_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"orbitgraphql/test_api/todo/db"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/rand"
)

type GraphQLClient struct {
	baseURL    string
	graphqlURL string
	headers    map[string]interface{}
}

func NewGraphQLClient(graphqlURL string, baseURL string, headers map[string]interface{}) *GraphQLClient {
	return &GraphQLClient{graphqlURL: graphqlURL, baseURL: baseURL, headers: headers}
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []interface{}   `json:"errors"`
}

func (c *GraphQLClient) MakeRequest(buf *bytes.Buffer, contentType string) (json.RawMessage, map[string]interface{}, time.Duration, error) {
	start := time.Now()

	responseHeaders := make(map[string]interface{})

	req, err := http.NewRequest("POST", c.graphqlURL, buf)
	if err != nil {
		return nil, responseHeaders, time.Since(start), err
	}

	req.Header.Set("Content-Type", contentType)
	for k, v := range c.headers {
		req.Header.Set(k, fmt.Sprintf("%v", v))
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, responseHeaders, time.Since(start), err
	}
	defer resp.Body.Close()

	for k := range resp.Header {
		responseHeaders[k] = resp.Header.Get(k)
	}

	res, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("received error in reading response ", err, resp.ContentLength)
		return nil, responseHeaders, time.Since(start), err
	}

	var response GraphQLResponse
	if err := json.Unmarshal(res, &response); err != nil {
		return nil, responseHeaders, time.Since(start), err
	}

	if len(response.Errors) > 0 {
		return nil, responseHeaders, time.Since(start), fmt.Errorf("graphql errors: %v", response.Errors)
	}

	return response.Data, responseHeaders, time.Since(start), nil
}

func (c *GraphQLClient) MakeJSONRequest(query string, variables map[string]interface{}) (json.RawMessage, map[string]interface{}, time.Duration, error) {
	start := time.Now()
	requestBody, err := json.Marshal(GraphQLRequest{Query: query, Variables: variables})
	if err != nil {
		return nil, nil, time.Since(start), err
	}

	res, headers, tt, err := c.MakeRequest(bytes.NewBuffer(requestBody), "application/json")
	return res, headers, tt, err
}

func (c *GraphQLClient) FlushCache() error {
	if c.baseURL == "" {
		return nil
	}
	resp, err := http.Post(c.baseURL+"/flush", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *GraphQLClient) FlushByType(typeName, id string) error {
	if c.baseURL == "" {
		return nil
	}
	requestBody, err := json.Marshal(map[string]interface{}{
		"type": typeName,
		"id":   id,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(c.baseURL+"/flush.type", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *GraphQLClient) CreateRandomUser() (*db.User, map[string]interface{}, time.Duration, error) {
	query := `
        mutation CreateUser($name: String!, $email: String!, $username: String!) {
            createUser(name: $name, email: $email, username: $username) {
                id
								name
								email
								username
								tags
								todosCount
								completionRate
								completionRateLast7Days
								activityStreak7Days
								meta {
									ipAddress
									userAgent
									createdEpoch
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"name":     fmt.Sprintf("User%d", rand.Intn(1000)),
		"email":    fmt.Sprintf("user%d@example.com", rand.Intn(1000)),
		"username": fmt.Sprintf("user%d", rand.Intn(1000)),
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		CreateUser db.User `json:"createUser"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.CreateUser, headers, time, nil
}

func (c *GraphQLClient) UpdateUser(id, name, email, username string) (*db.User, map[string]interface{}, time.Duration, error) {
	query := `
        mutation UpdateUser($id: String!, $name: String, $email: String, $username: String) {
            updateUser(id: $id, name: $name, email: $email, username: $username) {
                id
								name
								email
								username
								tags
								todosCount
								completionRate
								completionRateLast7Days
								activityStreak7Days
								meta {
									ipAddress
									userAgent
									createdEpoch
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id":       id,
		"name":     name,
		"username": username,
		"email":    email,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		UpdateUser db.User `json:"updateUser"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.UpdateUser, headers, time, nil
}

func (c *GraphQLClient) DeleteUser(userId string) (*db.User, map[string]interface{}, time.Duration, error) {
	query := `
        mutation DeleteUser($id: String!) {
            deleteUser(id: $id) {
                id
								name
								email
								username
								tags
								todosCount
								completionRate
								completionRateLast7Days
								activityStreak7Days
								meta {
									ipAddress
									userAgent
									createdEpoch
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": userId,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		DeleteUser db.User `json:"deleteUser"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.DeleteUser, headers, time, nil
}

func (c *GraphQLClient) CreateRandomTodo(userId string) (*db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        mutation CreateTodo($text: String!, $userId: String!) {
            createTodo(params: {text: $text, userId: $userId}) {
                id
								text
								done
								userId
								meta
								activityHistory
								user {
									id
									name
									email
									username
									tags
									todosCount
									completionRate
									completionRateLast7Days
									activityStreak7Days
									createdAt
									updatedAt
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"text":   fmt.Sprintf("Todo %d", rand.Intn(1000)),
		"userId": userId, // Replace with actual user ID
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		CreateTodo db.Todo `json:"createTodo"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.CreateTodo, headers, time, nil
}

func (c *GraphQLClient) MarkTodoAsDone(todoId string) (*db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        mutation MarkAsDone($id: String!) {
            markAsDone(id: $id) {
                id
								text
								done
								userId
								meta
								activityHistory
								user {
									id
									name
									email
									username
									todosCount
									completionRate
									completionRateLast7Days
									activityStreak7Days
									createdAt
									updatedAt
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": todoId, // Replace with actual user ID
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		MarkAsDone db.Todo `json:"markAsDone"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.MarkAsDone, headers, time, nil
}

func (c *GraphQLClient) MarkTodoAsUnDone(todoId string) (*db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        mutation MarkAsUndone($id: String!) {
            markAsUndone(id: $id) {
                id
								text
								done
								userId
								meta
								activityHistory
								user {
									id
									name
									email
									username
									todosCount
									completionRate
									completionRateLast7Days
									activityStreak7Days
									createdAt
									updatedAt
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": todoId, // Replace with actual user ID
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		MarkAsUndone db.Todo `json:"markAsUndone"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.MarkAsUndone, headers, time, nil
}

func (c *GraphQLClient) DeleteTodo(todoId string) (*db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        mutation DeleteTodo($id: String!) {
            deleteTodo(id: $id) {
                id
								text
								done
								userId
								meta
								activityHistory
								user {
									id
									name
									email
									username
									todosCount
									completionRate
									completionRateLast7Days
									activityStreak7Days
									createdAt
									updatedAt
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": todoId, // Replace with actual user ID
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		DeleteTodo db.Todo `json:"deleteTodo"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.DeleteTodo, headers, time, nil
}

func (c *GraphQLClient) PaginateUsers() ([]db.User, map[string]interface{}, time.Duration, error) {
	query := `
        query PaginateUsers($query: String $page: Int, $perPage: Int) {
            users(query: $query, page: $page, perPage: $perPage) {
                id
                name
                email
                username
								tags
								todosCount
								completionRate
								completionRateLast7Days
								activityStreak7Days
								meta {
									ipAddress
									userAgent
									createdEpoch
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"query":   "",
		"page":    1,
		"perPage": 100,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		Users []db.User `json:"users"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return result.Users, headers, time, nil
}

func (c *GraphQLClient) PaginateTodos() ([]db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        query PaginateTodos($query: String $page: Int, $perPage: Int) {
            todos(query: $query, page: $page, perPage: $perPage) {
                id
                text
                done
                userId
								createdAt
								updatedAt
								meta
								activityHistory
            }
        }
    `
	variables := map[string]interface{}{
		"query":   "",
		"page":    1,
		"perPage": 100,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		Todos []db.Todo `json:"todos"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return result.Todos, headers, time, nil
}

func (c *GraphQLClient) GetUserByID(userID string) (*db.User, map[string]interface{}, time.Duration, error) {
	query := `
        query GetUserByID($id: String!) {
            user(id: $id) {
                id
                name
                email
                username
								tags
								todosCount
								completionRate
								completionRateLast7Days
								activityStreak7Days
								meta {
									ipAddress
									userAgent
									createdEpoch
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": userID,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		User db.User `json:"user"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.User, headers, time, nil
}

func (c *GraphQLClient) GetUserTodosByID(userID string) (*db.User, map[string]interface{}, time.Duration, error) {
	query := `
        query GetUserTodosByID($id: String!) {
            user(id: $id) {
                id
                name
                email
                username
								tags
								todosCount
								completionRate
								completionRateLast7Days
								activityStreak7Days
								todos {
									id
									text
									done
									userId
									createdAt
									updatedAt
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": userID,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		User db.User `json:"user"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.User, headers, time, nil
}

func (c *GraphQLClient) GetTodoByID(todoID string) (*db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        query GetTodoByID($id: String!) {
            todo(id: $id) {
                id
                text
                done
                userId
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": todoID,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		Todo db.Todo `json:"todo"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.Todo, headers, time, nil
}

func (c *GraphQLClient) GetTodoByIDWithUser(todoID string) (*db.Todo, map[string]interface{}, time.Duration, error) {
	query := `
        query GetTodoByID($id: String!) {
            todo(id: $id) {
                id
                text
                done
                userId
								user {
									id
									name
								}
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": todoID,
	}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result struct {
		Todo db.Todo `json:"todo"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return &result.Todo, headers, time, nil
}

func (c *GraphQLClient) GetSystemDetails() (map[string]interface{}, map[string]interface{}, time.Duration, error) {
	query := `
        query GetSystemDetails {
						healthy
						totalTodos
						activityStreak7Days
						completionRateLast7Days
						completionRate
						activityHistory
						meta
        }
    `
	variables := map[string]interface{}{}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return nil, headers, time, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, headers, time, err
	}

	return result, headers, time, nil
}

func (c *GraphQLClient) DeleteEverything() (bool, map[string]interface{}, time.Duration, error) {

	// return true, 0, nil
	query := `
	      mutation DeleteEverything {
	          deleteEverything
	      }
	  `
	variables := map[string]interface{}{}

	data, headers, time, err := c.MakeJSONRequest(query, variables)
	if err != nil {
		return false, headers, time, err
	}

	var result struct {
		DeleteEverything bool `json:"deleteEverything"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, headers, time, err
	}

	return result.DeleteEverything, headers, time, nil
}

func (c *GraphQLClient) UploadImage(filePath string) (map[string]interface{}, map[string]interface{}, time.Duration, error) {
	startTime := time.Now()

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("operations", "{ \"query\": \"mutation ($file: Upload!) { uploadImage(file: $file) { base64 mimeType } }\", \"variables\": { \"file\": null } }")
	_ = writer.WriteField("map", "{ \"0\": [\"variables.file\"] }")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println(err)
		return nil, nil, time.Since(startTime), err
	}
	defer file.Close()
	part3, err := writer.CreateFormFile("0", filepath.Base(filePath))
	if err != nil {
		fmt.Println(err)
		return nil, nil, time.Since(startTime), err
	}
	_, err = io.Copy(part3, file)
	if err != nil {
		fmt.Println(err)
		return nil, nil, time.Since(startTime), err
	}
	err = writer.Close()
	if err != nil {
		fmt.Println(err)
		return nil, nil, time.Since(startTime), err
	}

	res, headers, tt, err := c.MakeRequest(payload, writer.FormDataContentType())
	if err != nil {
		return nil, headers, tt, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(res, &result); err != nil {
		return nil, headers, tt, err
	}

	return result, headers, tt, nil
}
