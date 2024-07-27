package main

import (
	"fmt"
	"graphql_cache/test_api/todo/db"
	"graphql_cache/utils/test_endpoints"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const NUMBER_OF_USERS = 10

func RunUsersOperations(t *testing.T, client *test_endpoints.GraphQLClient) time.Duration {
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
	client := test_endpoints.NewGraphQLClient("http://localhost:9090/graphql", "http://localhost:9090")
	defer client.DeleteEverything()
	defer client.FlushCache()
	timeTaken := RunUsersOperations(t, client)
	fmt.Printf("Total time taken for Cached Todo API: %v\n", timeTaken)
}

func TestAPIDefaultTestSuite(t *testing.T) {
	client := test_endpoints.NewGraphQLClient("http://localhost:8080/graphql", "")
	defer client.DeleteEverything()
	defer client.FlushCache()
	timeTaken := RunUsersOperations(t, client)
	fmt.Printf("Total time taken for Default Todo API: %v\n", timeTaken)
}
