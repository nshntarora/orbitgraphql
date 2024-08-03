package main

import (
	"context"
	"fmt"
	"graphql_cache/test_api/todo/db"
	"graphql_cache/test_api/todo/graph"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func main() {
	fmt.Println(`
___  __   __   __  
 |  /  \ |  \ /  \ 
 |  \__/ |__/ \__/ 
                   
Todo List GraphQL Server for graph_cache tests
	`)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	conn := db.Setup()
	defer db.Close(conn.DB)

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{DB: &conn}}))

	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		time.Sleep(100 * time.Millisecond)
		response := next(ctx)
		return response
	})

	http.Handle("/", playground.Handler("GraphQL playground", "/graphql"))
	http.Handle("/graphql", srv)

	fmt.Printf("connect to http://127.0.0.1:%s/ for GraphQL playground", port)
	fmt.Println("\ngraphql API is available on http://127.0.0.1:" + port + "/graphql")
	log.Fatal(http.ListenAndServe("127.0.0.1:"+port, nil))
}
