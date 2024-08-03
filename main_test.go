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

func RunUsersOperations(t *testing.T, client *test_client.GraphQLClient) time.Duration {
	totalTimeTaken := time.Duration(0)

	deleted, tt, err := client.DeleteEverything()
	if err != nil {
		fmt.Println(err)
		return totalTimeTaken
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
			fmt.Println(err)
			return totalTimeTaken
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
			fmt.Println(err)
			return totalTimeTaken
		}
		totalTimeTaken += tt

		assert.Equal(t, NUMBER_OF_USERS, len(users))
		// assert.Equal(t, createdUsers, users)
	}

	user, tt, err := client.CreateRandomUser()
	if err != nil {
		fmt.Println(err)
		return totalTimeTaken
	}
	totalTimeTaken += tt

	client.FlushByType("User", "")

	users, tt, err := client.PaginateUsers()
	if err != nil {
		fmt.Println(err)
		return totalTimeTaken
	}
	totalTimeTaken += tt

	assert.Equal(t, NUMBER_OF_USERS+1, len(users))

	for i := 0; i < NUMBER_OF_USERS; i++ {

		user, tt, err = client.GetUserByID(userIDToUpdate)
		if err != nil {
			fmt.Println(err)
			return totalTimeTaken
		}
		totalTimeTaken += tt

		assert.NotNil(t, user)
		assert.NotEqual(t, "Updated Name", user.Name)
	}

	// Update a user
	user, tt, err = client.UpdateUser(userIDToUpdate, "Updated Name", "", "")
	if err != nil {
		fmt.Println(err)
		return totalTimeTaken
	}
	totalTimeTaken += tt

	assert.Equal(t, "Updated Name", user.Name)

	for i := 0; i < NUMBER_OF_USERS; i++ {
		updatedUser, tt, err := client.GetUserByID(userIDToUpdate)
		if err != nil {
			fmt.Println(err)
			return totalTimeTaken
		}
		totalTimeTaken += tt

		assert.Equal(t, "Updated Name", updatedUser.Name)
	}

	return totalTimeTaken

}

func TestAPICacheTestSuite(t *testing.T) {
	client := test_client.NewGraphQLClient("http://localhost:9090/graphql", "http://localhost:9090")
	defer client.DeleteEverything()
	defer client.FlushCache()
	timeTaken := RunUsersOperations(t, client)
	fmt.Printf("Total time taken for Cached Todo API: %v\n", timeTaken)
}

func TestAPIDefaultTestSuite(t *testing.T) {
	client := test_client.NewGraphQLClient("http://localhost:8080/graphql", "")
	defer client.DeleteEverything()
	defer client.FlushCache()
	timeTaken := RunUsersOperations(t, client)
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
		if err != nil {
			fmt.Println(err)
			return
		}
		createdUsers = append(createdUsers, *user)
	}

	cachedUsersResponses := []db.User{}

	timeTakenForCacheClient := time.Duration(0)

	for i := 0; i < NUMBER_OF_USERS; i++ {
		users, tt, err := cacheClient.PaginateUsers()
		if err != nil {
			fmt.Println(err)
			return
		}
		cachedUsersResponses = append(cachedUsersResponses, users...)
		timeTakenForCacheClient += tt
	}

	defaultUserResponses := []db.User{}
	timeTakenForDefaultClient := time.Duration(0)
	for i := 0; i < NUMBER_OF_USERS; i++ {
		users, tt, err := defaultClient.PaginateUsers()
		if err != nil {
			fmt.Println(err)
			return
		}
		defaultUserResponses = append(defaultUserResponses, users...)
		timeTakenForDefaultClient += tt
	}

	fmt.Println("Time taken for cached API: ", timeTakenForCacheClient)
	fmt.Println("Time taken for default API: ", timeTakenForDefaultClient)

	assert.Equal(t, len(cachedUsersResponses), len(defaultUserResponses))
	assert.Equal(t, cachedUsersResponses, defaultUserResponses)
}
