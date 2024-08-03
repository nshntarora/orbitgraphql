package main

import (
	"fmt"
	"graphql_cache/api"
	"graphql_cache/config"
	"log"
	"strconv"
)

func main() {

	cfg := config.NewConfig()

	server := api.NewServer(cfg)

	// Start the server and log any errors
	fmt.Println("starting server on :" + strconv.Itoa(cfg.Port))
	err := server.Start()
	if err != nil {
		log.Fatal("error starting server: ", err)
	}
}
