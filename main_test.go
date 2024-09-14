package main

import (
	"fmt"
	"graphql_cache/test_api/todo/db"
	"graphql_cache/test_client"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const NUMBER_OF_USERS = 10

func RunUsersOperations(t *testing.T, client *test_client.GraphQLClient) (time.Duration, error) {
	totalTimeTaken := time.Duration(0)

	deleted, tt, err := client.DeleteEverything()
	if err != nil {
		fmt.Println("error resetting database ", err)
		return totalTimeTaken, err
	}

	totalTimeTaken += tt

	client.FlushCache()

	assert.Equal(t, deleted, true)

	userIDToUpdate := ""
	// Create 5 users
	createdUsers := []db.User{}
	for i := 0; i < NUMBER_OF_USERS; i++ {
		user, tt, err := client.CreateRandomUser()
		if err != nil {
			fmt.Println("error creating user ", err)
			return totalTimeTaken, err
		}
		totalTimeTaken += tt
		userIDToUpdate = user.ID.String()
		createdUsers = append(createdUsers, *user)
	}

	assert.NotNil(t, userIDToUpdate)

	// Paginate users
	for i := 0; i < NUMBER_OF_USERS; i++ {
		users, tt, err := client.PaginateUsers()
		if err != nil {
			fmt.Println("error paginating users ", err)
			return totalTimeTaken, err
		}
		totalTimeTaken += tt

		assert.Equal(t, NUMBER_OF_USERS, len(users))
		assert.Equal(t, createdUsers, users)
	}

	user, tt, err := client.CreateRandomUser()
	if err != nil {
		fmt.Println("error creating user ", err)
		return totalTimeTaken, err
	}
	assert.NotNil(t, user)
	totalTimeTaken += tt

	client.FlushByType("User", "")

	users, tt, err := client.PaginateUsers()
	if err != nil {
		fmt.Println("error paginating users ", err)
		return totalTimeTaken, err
	}
	totalTimeTaken += tt

	assert.Equal(t, NUMBER_OF_USERS+1, len(users))

	for i := 0; i < NUMBER_OF_USERS; i++ {

		user, tt, err = client.GetUserByID(userIDToUpdate)
		if err != nil {
			fmt.Println("error fetching user", err)
			return totalTimeTaken, err
		}
		totalTimeTaken += tt

		assert.NotNil(t, user)
		assert.NotEqual(t, "Updated Name", user.Name)
	}

	// Update a user
	user, tt, err = client.UpdateUser(userIDToUpdate, "Updated Name", "", "")
	if err != nil {
		fmt.Println("error updating user ", err)
		return totalTimeTaken, err
	}
	totalTimeTaken += tt

	assert.Equal(t, "Updated Name", user.Name)

	for i := 0; i < NUMBER_OF_USERS; i++ {
		updatedUser, tt, err := client.GetUserByID(userIDToUpdate)
		if err != nil {
			fmt.Println("error fetching user ", err)
			return totalTimeTaken, err
		}
		assert.Greater(t, len(updatedUser.Tags), 1)
		assert.Contains(t, updatedUser.Meta, "ipAddress")
		totalTimeTaken += tt

		assert.Equal(t, "Updated Name", updatedUser.Name)
	}

	return totalTimeTaken, nil

}

func TestAPICacheTestSuite(t *testing.T) {
	client := test_client.NewGraphQLClient("http://localhost:9090/graphql", "http://localhost:9090")
	defer client.DeleteEverything()
	defer client.FlushCache()
	timeTaken, err := RunUsersOperations(t, client)
	assert.Nil(t, err)
	fmt.Printf("Total time taken for Cached Todo API: %v\n", timeTaken)
}

func TestAPIDefaultTestSuite(t *testing.T) {
	client := test_client.NewGraphQLClient("http://localhost:8080/graphql", "")
	defer client.DeleteEverything()
	defer client.FlushCache()
	timeTaken, err := RunUsersOperations(t, client)
	assert.Nil(t, err)
	fmt.Printf("Total time taken for Default Todo API: %v\n", timeTaken)
}

func TestAPIResponseConsistency(t *testing.T) {
	cacheClient := test_client.NewGraphQLClient("http://localhost:9090/graphql", "http://localhost:9090")
	cacheClient.DeleteEverything()
	cacheClient.FlushCache()

	defaultClient := test_client.NewGraphQLClient("http://localhost:8080/graphql", "")

	// run all mutations, then run all queries on the cached api and the default API, then compare the responses. If the responses do not match, the test fails.

	// Create 5 users
	createdUsers := []db.User{}
	for i := 0; i < NUMBER_OF_USERS; i++ {
		user, _, err := cacheClient.CreateRandomUser()
		assert.Nil(t, err)
		createdUsers = append(createdUsers, *user)
	}

	cachedUsersResponses := []db.User{}

	timeTakenForCacheClient := time.Duration(0)

	for i := 0; i < NUMBER_OF_USERS; i++ {
		users, tt, err := cacheClient.PaginateUsers()
		assert.Nil(t, err)
		cachedUsersResponses = append(cachedUsersResponses, users...)
		timeTakenForCacheClient += tt
	}

	defaultUserResponses := []db.User{}
	timeTakenForDefaultClient := time.Duration(0)
	for i := 0; i < NUMBER_OF_USERS; i++ {
		users, tt, err := defaultClient.PaginateUsers()
		assert.Nil(t, err)
		defaultUserResponses = append(defaultUserResponses, users...)
		timeTakenForDefaultClient += tt
	}

	fmt.Println("Time taken for cached API: ", timeTakenForCacheClient)
	fmt.Println("Time taken for default API: ", timeTakenForDefaultClient)

	assert.Equal(t, len(cachedUsersResponses), len(defaultUserResponses))
	assert.Equal(t, cachedUsersResponses, defaultUserResponses)

	cachedSystemDetailsResponses := []map[string]interface{}{}
	defaultSystemDetailsResponses := []map[string]interface{}{}
	// get system details from cache client and default client
	for i := 0; i < NUMBER_OF_USERS; i++ {
		details, _, err := defaultClient.GetSystemDetails()
		assert.NotNil(t, details)
		assert.Nil(t, err)
		defaultSystemDetailsResponses = append(defaultSystemDetailsResponses, details)
	}

	// get system details from cache client and default client
	for i := 0; i < NUMBER_OF_USERS; i++ {
		details, _, err := cacheClient.GetSystemDetails()
		assert.NotNil(t, details)
		assert.Nil(t, err)
		cachedSystemDetailsResponses = append(cachedSystemDetailsResponses, details)
	}

	assert.Equal(t, len(cachedSystemDetailsResponses), len(defaultSystemDetailsResponses))
	assert.Equal(t, cachedSystemDetailsResponses, defaultSystemDetailsResponses)
}

func TestWithTodoOperations(t *testing.T) {
	client := test_client.NewGraphQLClient("http://localhost:9090/graphql", "http://localhost:9090")
	defer client.DeleteEverything()
	defer client.FlushCache()

	user, _, err := client.CreateRandomUser()
	assert.Nil(t, err)

	userID := user.ID.String()

	for i := 0; i < 10; i++ {
		todo, _, err := client.CreateRandomTodo(userID)
		assert.Nil(t, err)
		assert.NotNil(t, todo)
	}

	todo, _, err := client.CreateRandomTodo(userID)
	assert.Nil(t, err)
	assert.NotNil(t, todo)

	todoID := todo.ID.String()

	// Get the todo
	todoResponse, _, err := client.GetTodoByID(todoID)
	assert.Nil(t, err)
	assert.NotNil(t, todoResponse)

	assert.NotNil(t, todoResponse)
	assert.Equal(t, todo.ID, todoResponse.ID)
	assert.Equal(t, todo.Text, todoResponse.Text)
	assert.Equal(t, todo.Done, todoResponse.Done)

	// Update the todo
	todoResponse, _, err = client.MarkTodoAsDone(todoID)
	assert.Nil(t, err)
	assert.NotNil(t, todoResponse)

	assert.NotNil(t, todoResponse)
	assert.Equal(t, todo.ID, todoResponse.ID)
	assert.Equal(t, todo.Text, todoResponse.Text)
	assert.Equal(t, true, todoResponse.Done)

	// get the todo again
	todoResponse, _, err = client.GetTodoByID(todoID)
	assert.Nil(t, err)
	assert.NotNil(t, todoResponse)

	assert.NotNil(t, todoResponse)
	assert.Equal(t, todo.ID, todoResponse.ID)
	assert.Equal(t, todo.Text, todoResponse.Text)
	assert.Equal(t, true, todoResponse.Done)
	assert.Nil(t, todoResponse.User)

	// get the todo with user
	todoResponse, _, err = client.GetTodoByIDWithUser(todoID)
	assert.Nil(t, err)
	assert.NotNil(t, todoResponse)

	assert.NotNil(t, todoResponse)
	assert.Equal(t, todo.ID, todoResponse.ID)
	assert.Equal(t, todo.Text, todoResponse.Text)
	assert.Equal(t, true, todoResponse.Done)
	assert.NotNil(t, todoResponse.User)
	assert.Equal(t, userID, todoResponse.User.ID.String())
}

func TestMultipartRequestsBypass(t *testing.T) {
	client := test_client.NewGraphQLClient("http://localhost:9090/graphql", "http://localhost:9090")

	resp, headers, _, err := client.UploadImage("./logo.svg")
	assert.Nil(t, err)
	assert.NotNil(t, resp)
	assert.NotNil(t, headers)

	assert.Equal(t, "BYPASS", headers["X-Orbit-Cache"])
}
