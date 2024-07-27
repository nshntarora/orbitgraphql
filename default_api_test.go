package main

import (
	"fmt"
	"graphql_cache/test_api/todo/db"
	"graphql_cache/utils/test_endpoints"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCachedTodoAPI(t *testing.T) {
	// 	1. Create 5 users with random name, email, and username
	// 2. Paginate users and see if the 5 are there
	// 3. Update a user's name with the another random name
	// 4. Get the updated user's id, and get that user to see if the name is updated

	client := test_endpoints.NewGraphQLClient("http://localhost:9090/graphql")

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
	createdUsers := []db.User{}
	for i := 0; i < 5; i++ {
		user, tt, err := client.CreateRandomUser()
		if err != nil {
			fmt.Println(err)
			return
		}
		totalTimeTaken += tt
		userIDToUpdate = user.ID.String()
		createdUsers = append(createdUsers, *user)
	}

	assert.NotNil(t, userIDToUpdate)

	// Paginate users
	for i := 0; i < 5; i++ {
		users, tt, err := client.PaginateUsers()
		if err != nil {
			fmt.Println(err)
			return
		}
		totalTimeTaken += tt

		assert.Equal(t, 5, len(users))
		assert.Equal(t, createdUsers, users)
	}

	_, tt, err = client.CreateRandomUser()
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	users, tt, err := client.PaginateUsers()
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, 6, len(users))

	_, tt, err = client.GetUserByID(userIDToUpdate)
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	// Update a user
	user, tt, err := client.UpdateUser(userIDToUpdate, "Updated Name", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, "Updated Name", user.Name)

	for i := 0; i < 5; i++ {
		updatedUser, tt, err := client.GetUserByID(userIDToUpdate)
		if err != nil {
			fmt.Println(err)
			return
		}
		totalTimeTaken += tt

		assert.Equal(t, "Updated Name", updatedUser.Name)
	}

	fmt.Printf("Total time taken for Cached Todo API: %v\n", totalTimeTaken)
}

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

	for i := 0; i < 5; i++ {
		users, tt, err := client.PaginateUsers()
		if err != nil {
			fmt.Println(err)
			return
		}
		totalTimeTaken += tt

		assert.Equal(t, 5, len(users))
	}

	_, tt, err = client.CreateRandomUser()
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	users, tt, err := client.PaginateUsers()
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, 6, len(users))

	_, tt, err = client.GetUserByID(userIDToUpdate)
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	// Update a user
	user, tt, err := client.UpdateUser(userIDToUpdate, "Updated Name", "", "")
	if err != nil {
		fmt.Println(err)
		return
	}
	totalTimeTaken += tt

	assert.Equal(t, "Updated Name", user.Name)

	for i := 0; i < 5; i++ {
		updatedUser, tt, err := client.GetUserByID(userIDToUpdate)
		if err != nil {
			fmt.Println(err)
			return
		}
		totalTimeTaken += tt

		assert.Equal(t, "Updated Name", updatedUser.Name)
	}

	fmt.Printf("Total time taken for Default Todo API: %v\n", totalTimeTaken)
}
