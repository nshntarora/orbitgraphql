package main

import "fmt"

func main() {
	fmt.Println("👋 Hi! Want to run GraphQL cache? It is only available as a package.")
	fmt.Println("While building, we're using tests to see the impact of the cache.")
	fmt.Println("\ngo test ./... -v -count=1")
	fmt.Println("\nFor more details on how to use the package, please refer to the README.md file.")
}
