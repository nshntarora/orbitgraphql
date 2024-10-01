---
description: Get it running locally on your machine and see how it works!
icon: bullseye-arrow
---

# Quickstart

### Prerequisites

To build the project locally, you need to have `Go` installed on your machine - [_Installing Go_](https://go.dev/doc/install)

### Running

Once you have the project cloned,

1. Run the following command in the project directory to install/download all dependencies

```
go mod tidy
```

2. Update the configuration in the `config.toml` file (add your origin URL, the port, etc.)
3. Run the command below to start the cache server

```
go run main.go
```

4. That's it! You can start making requests to your cache server.

### Building

You can build the executable for the server by running the command below

```
go build -o orbitgraphql main.go
```

This will build the binary on your machine.

Please note, the build binary will need the `config.toml` file in the same directory if you're providing configuration in the file. You can also provide configuration for the server using envionment variables.

### Docker

First thing you need to do is run the server. You can run it using Docker.

#### **Docker Build**

```sh
docker build . -t orbitgraphql
```

#### **Running the Docker container**

```sh
docker run -p 9090:9090 -e ORBIT_ORIGIN=http://localhost:8080/graphql orbitgraphql
```

All requests to `localhost:9090/` will be proxied to `localhost:8080/graphql` and requests which result in a cache `HIT` will be served directly.
