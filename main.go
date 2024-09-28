package main

import (
	"fmt"
	"log"
	"orbitgraphql/api"
	"orbitgraphql/config"
	"orbitgraphql/logger"
	"strconv"
)

func main() {

	fmt.Println(`	
   ____     ___      _    _ _      ___               _    ___  _    
  / /\ \   / _ \ _ _| |__(_) |_   / __|_ _ __ _ _ __| |_ / _ \| |   
 | |  | | | (_) | '_| '_ \ |  _| | (_ | '_/ _' | '_ \ ' \ (_) | |__ 
 | |  | |  \___/|_| |_.__/_|\__|  \___|_| \__,_| .__/_||_\__\_\____|
  \_\/_/                                       |_|                  
	`)

	fmt.Println("⏳ starting server...")
	fmt.Println("🛠️ initializing configuration...")
	cfg := config.NewConfig()
	fmt.Println("🛠️ configuration initalized")
	fmt.Println("⚙️ configuration: ")
	fmt.Print("→ cache_backend=", cfg.CacheBackend, "\n→ cache_header_name=", cfg.CacheHeaderName, "\n→ origin=", cfg.Origin, "\n→ port=", cfg.Port, "\n→ scope_headers=", cfg.ScopeHeaders, "\n→ primary_key_field=", cfg.PrimaryKeyField, "\n→ log_level=", cfg.LogLevel, "\n→ log_format=", cfg.LogFormat, "\n→ redis_host=", cfg.RedisHost, "\n→ redis_port=", cfg.RedisPort, "\n→ cache_ttl=", cfg.CacheTTL, "\n→ handlers_graphql_path=", cfg.HandlersGraphQLPath, "\n→ handlers_flush_all_path=", cfg.HandlersFlushAllPath, "\n→ handlers_flush_by_type_path=", cfg.HandlersFlushByTypePath, "\n→ handlers_debug_path=", cfg.HandlersDebugPath, "\n→ handlers_health_path=", cfg.HandlersHealthPath, "\n\n")

	logger.Configure(&logger.Config{
		Format: string(cfg.LogFormat),
		Level:  cfg.LogLevel,
	})

	server := api.NewServer(cfg)

	// Start the server and log any errors
	fmt.Println("✅ server started on port :" + strconv.Itoa(cfg.Port))
	err := server.Start()
	if err != nil {
		log.Fatal("‼️ error starting server: ", err)
	}
}
