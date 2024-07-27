package main

import (
	"fmt"
	"graphql_cache/utils/test_endpoints"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTodoAPI(t *testing.T) {
	// 	1. Create 5 users with random name, email, and username
	// 2. Paginate users and see if the 5 are there
	// 3. Update a user's name with the another random name
	// 4. Get the updated user's id, and get that user to see if the name is updated

	client := test_endpoints.NewGraphQLClient("http://localhost:8080/graphql")

	totalTimeTaken := time.Duration(0)

	deleted, tt, err := client.DeleteEverything()
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, deleted, true)

	userIDToUpdate := ""
	// Create 5 users
	for i := 0; i < 5; i++ {
		user, tt, err := client.CreateRandomUser()
		if err != nil {
			fmt.Println(err)
			return
		}
		totalTimeTaken += tt

		userIDToUpdate = user.ID.String()
	}

	assert.NotNil(t, userIDToUpdate)

	// Paginate users
	users, tt, err := client.PaginateUsers()
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, len(users), 5)

	// Update a user
	user, tt, err := client.UpdateUser(userIDToUpdate, "Updated Name", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, user.Name, "Updated Name")

	// Get the updated user
	updatedUser, tt, err := client.GetUserByID(userIDToUpdate)
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, updatedUser.Name, "Updated Name")

	fmt.Printf("Total time taken for Default Todo API: %v\n", totalTimeTaken)
}
