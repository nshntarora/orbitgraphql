package test_client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"graphql_cache/test_api/todo/db"
	"net/http"
	"time"

	"golang.org/x/exp/rand"
)

type GraphQLClient struct {
	baseURL    string
	graphqlURL string
}

func NewGraphQLClient(graphqlURL string, baseURL string) *GraphQLClient {
	return &GraphQLClient{graphqlURL: graphqlURL, baseURL: baseURL}
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []interface{}   `json:"errors"`
}

func (c *GraphQLClient) MakeRequest(query string, variables map[string]interface{}) (json.RawMessage, time.Duration, error) {
	start := time.Now()
	requestBody, err := json.Marshal(GraphQLRequest{Query: query, Variables: variables})

	if err != nil {
		return nil, time.Since(start), err
	}

	resp, err := http.Post(c.graphqlURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, time.Since(start), err
	}
	defer resp.Body.Close()

	var response GraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, time.Since(start), err
	}

	if len(response.Errors) > 0 {
		return nil, time.Since(start), fmt.Errorf("graphql errors: %v", response.Errors)
	}

	return response.Data, time.Since(start), nil
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

func (c *GraphQLClient) CreateRandomUser() (*db.User, time.Duration, error) {
	query := `
        mutation CreateUser($name: String!, $email: String!, $username: String!) {
            createUser(name: $name, email: $email, username: $username) {
                id
								name
								email
								username
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

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		CreateUser db.User `json:"createUser"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return &result.CreateUser, time, nil
}

func (c *GraphQLClient) UpdateUser(id, name, email, username string) (*db.User, time.Duration, error) {
	query := `
        mutation UpdateUser($id: String!, $name: String, $email: String, $username: String) {
            updateUser(id: $id, name: $name, email: $email, username: $username) {
                id
								name
								email
								username
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

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		UpdateUser db.User `json:"updateUser"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return &result.UpdateUser, time, nil
}

func (c *GraphQLClient) CreateRandomTodo(userId string) (*db.Todo, time.Duration, error) {
	query := `
        mutation CreateTodo($text: String!, $userId: String!) {
            createTodo(params: {text: $text, userId: $userId}) {
                id
								text
								done
								userId
								user {
									id
									name
									email
									username
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

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		CreateTodo db.Todo `json:"createTodo"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return &result.CreateTodo, time, nil
}

func (c *GraphQLClient) PaginateUsers() ([]db.User, time.Duration, error) {
	query := `
        query PaginateUsers($query: String $page: Int, $perPage: Int) {
            users(query: $query, page: $page, perPage: $perPage) {
                id
                name
                email
                username
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

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		Users []db.User `json:"users"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return result.Users, time, nil
}

func (c *GraphQLClient) PaginateTodos() ([]db.Todo, time.Duration, error) {
	query := `
        query PaginateTodos($query: String $page: Int, $perPage: Int) {
            todos(query: $query, page: $page, perPage: $perPage) {
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
		"query":   "",
		"page":    1,
		"perPage": 100,
	}

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		Todos []db.Todo `json:"todos"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return result.Todos, time, nil
}

func (c *GraphQLClient) GetUserByID(userID string) (*db.User, time.Duration, error) {
	query := `
        query GetUserByID($id: String!) {
            user(id: $id) {
                id
                name
                email
                username
								createdAt
								updatedAt
            }
        }
    `
	variables := map[string]interface{}{
		"id": userID,
	}

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		User db.User `json:"user"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return &result.User, time, nil
}

func (c *GraphQLClient) GetTodoByID(todoID string) (*db.Todo, time.Duration, error) {
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

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return nil, time, err
	}

	var result struct {
		Todo db.Todo `json:"todo"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, time, err
	}

	return &result.Todo, time, nil
}

func (c *GraphQLClient) DeleteEverything() (bool, time.Duration, error) {
	query := `
        mutation DeleteEverything {
            deleteEverything
        }
    `
	variables := map[string]interface{}{}

	data, time, err := c.MakeRequest(query, variables)
	if err != nil {
		return false, time, err
	}

	var result struct {
		DeleteEverything bool `json:"deleteEverything"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return false, time, err
	}

	return result.DeleteEverything, time, nil
}
