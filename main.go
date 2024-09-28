package main

import (
	"fmt"
	"graphql_cache/api"
	"graphql_cache/config"
	"graphql_cache/logger"
	"log"
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
	fmt.Println("⚙️ configuration: ", "cache_backend=", cfg.CacheBackend, "cache_header_name=", cfg.CacheHeaderName, "origin=", cfg.Origin, "port=", cfg.Port, "scope_headers=", cfg.ScopeHeaders, "primary_key_field=", cfg.PrimaryKeyField, "log_level=", cfg.LogLevel, "log_format=", cfg.LogFormat, "redis_host=", cfg.RedisHost, "redis_port=", cfg.RedisPort, "cache_ttl=", cfg.CacheTTL, "handlers_graphql_path=", cfg.HandlersGraphQLPath, "handlers_flush_all_path=", cfg.HandlersFlushAllPath, "handlers_flush_by_type_path=", cfg.HandlersFlushByTypePath, "handlers_debug_path=", cfg.HandlersDebugPath, "handlers_health_path=", cfg.HandlersHealthPath)

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
