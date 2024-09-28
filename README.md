![](./logo.svg)

# Orbit GraphQL

### Stellate and urql like cache for your GraphQL APIs

## Why?

Because all GraphQL requests are on a single endpoint and resources are differentiated based on what your request body looks like, we can't use HTTP caching methods for them (say goodbye to etags, 304s, etc.)

Currently the ecosystem solves for the problem two ways, client side cache or a server side cache. Orbit GraphQL is a solution to setting up a server side cache for your GraphQL API.

## Getting Started

### Docker

First thing you need to do is run the server. You can run it using Docker.

Docker Build

```
docker build . -t orbitgraphql
```

Running the Docker container

```
docker run -p 9090:9090 -e ORBIT_ORIGIN=http://localhost:8080/graphql orbitgraphql
```

All requests to `localhost:9090/` will be proxied to `localhost:8080/graphql` and requests which result in a cache `HIT` will be served directly from cache without hitting your origin server.

### Local Development

You can start the server on your machine locally by running:

```
go run main.go
```

## Configuration

You can configure the origin URL, the port the server runs on, the cache backend to use (currently in memory and redis are supported), and a few more things.

The system takes configuration in two ways,

1. The `config.toml` file in your project directory
2. Environment variables

The configuration is read from the `config.toml` file. You can also override the configuration using environment variables.

[View Configuration File](https://github.com/orbitgraphql/orbitgraphql/blob/main/config.toml)

## What's the current status of this project?

This is not production ready yet.

Here is a non-exhaustive list of things planned for the project:

1. Support for Fragments.
2. Benchmarking.
3. Go/JavaScript clients for the administration APIs (used to flush cache).
4. Better observability setup (to help monitor how the cache server is performing).
5. Support for analytics on top of your GraphQL API to help you get insights on how your API is being consumed.

## Is there a hosted version?

No. There's no plan to offer a hosted version for this as of now (but, never say never)
