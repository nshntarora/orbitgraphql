---
description: A list of things you can configure for the cache server
icon: screwdriver-wrench
---

# Configuration Options

You can configure the origin URL, the port the server runs on, the cache backend to use (currently in memory and redis are supported), and a few more things.

The system takes configuration in two ways,

1. The `config.toml` file in your project directory
2. Environment variables

The configuration is read from the `config.toml` file. You can also override the configuration using environment variables.

### Origin

The endpoint to which the cache should forward requests.

- **Configuration Key:** `origin`
- **Environment Variable:** `ORBIT_ORIGIN`
- **Default Value:** None. **Required**

### Port

The port that the cache will run on.

- **Configuration Key:** `port`
- **Environment Variable:** `ORBIT_PORT`
- **Default Value:** `9090`

### Cache Backend

The backend for caching values. Supported values are `redis` and `in_memory`. If you have cache backend configured as `redis` you will also need to provide Redis Host and Redis Port

- **Configuration Key:** `cache_backend`
- **Environment Variable:** `ORBIT_CACHE_BACKEND`
- **Default Value:** `"in_memory"`

### Redis Host

The host for the Redis cache backend.

- **Configuration Key:** `redis_host`
- **Environment Variable:** `ORBIT_REDIS_HOST`
- **Default Value:** `"localhost"`

### Redis Port

The port for the Redis cache backend.

- **Configuration Key:** `redis_port`
- **Environment Variable:** `ORBIT_REDIS_PORT`
- **Default Value:** `6379`

### Cache Header Name

The header name that returns cache status (`HIT`, `MISS`, or `BYPASS`).

- **Configuration Key:** `cache_header_name`
- **Environment Variable:** `ORBIT_CACHE_HEADER_NAME`
- **Default Value:** `"X-Orbit-Cache"`

### Cache TTL

The TTL (Time To Live) of the GraphQL cache in seconds. Default is 60 minutes.

- **Configuration Key:** `cache_ttl`
- **Environment Variable:** `ORBIT_CACHE_TTL`
- **Default Value:** `3600`

### Scope Headers

Headers used to scope the cache based on their unique values. To pass muliple headers add them as a comma separated string (example: `Authorization,X-API-Key`)

- **Configuration Key:** `scope_headers`
- **Environment Variable:** `ORBIT_SCOPE_HEADERS`
- **Default Value:** `"Authorization"`

### Primary Key Field

The field in GraphQL responses used to identify unique objects (this should be unique for every resource). Defaults to `id`.

- **Configuration Key:** `primary_key_field`
- **Environment Variable:** `ORBIT_PRIMARY_KEY_FIELD`
- **Default Value:** `"id"`

### Handlers GraphQL Path

The API path for GraphQL requests.

- **Configuration Key:** `handlers_graphql_path`
- **Environment Variable:** `ORBIT_HANDLERS_GRAPHQL_PATH`
- **Default Value:** `"/graphql"`

### Handlers Flush All Path

The API path to flush all cache.

- **Configuration Key:** `handlers_flush_all_path`
- **Environment Variable:** `ORBIT_HANDLERS_FLUSH_ALL_PATH`
- **Default Value:** `"/flush"`

### Handlers Flush By Type Path

The API path to flush cache by type.

- **Configuration Key:** `handlers_flush_by_type_path`
- **Environment Variable:** `ORBIT_HANDLERS_FLUSH_BY_TYPE_PATH`
- **Default Value:** `"/flush.type"`

### Handlers Debug Path

The API path for debugging.

- **Configuration Key:** `handlers_debug_path`
- **Environment Variable:** `ORBIT_HANDLERS_DEBUG_PATH`
- **Default Value:** `"/debug"`

### Handlers Health Path

The API path for health checks.

- **Configuration Key:** `handlers_health_path`
- **Environment Variable:** `ORBIT_HANDLERS_HEALTH_PATH`
- **Default Value:** `"/health"`

### Log Level

The level of logging. Supported values are `debug`, `info`, `warn`, `error`. Defaults to `info`.

- **Configuration Key:** `log_level`
- **Environment Variable:** `ORBIT_LOG_LEVEL`
- **Default Value:** `"info"`

### Log Format

The format for system logs. Supported values are `json` and `text`. Defaults to `text`.

- **Configuration Key:** `log_format`
- **Environment Variable:** `ORBIT_LOG_FORMAT`
- **Default Value:** `"text"`
