![](./logo.svg)

# Orbit GraphQL

### Stellate and urql like cache for your GraphQL APIs

## Why?

Because all GraphQL requests are on a single endpoint and resources are differentiated based on what your request body looks like, we can't use HTTP caching methods for them (say goodbye to etags, 304s, etc.)

## What does it do?

This package takes a GraphQL request, caches the response - if the same request is repeated again, we can serve the response directly from cache without hitting the origin server.

## Why can't I just use a key value store for caching my GraphQL requests?

You can, but the challenge comes with cache invalidation. How do you identify what fields have changed or have been added so you can invalidate your cache? One way is to invalidate everything for the user - but if you have updates happening often, it becomes as good as no cache.

Since all requests to your GraphQL server pass through `graphql_cache` (including mutations), it can automatically invalidate objects and entire queries for which data has changed. Basically, it supports partial cache invalidation.

## What's the current status of this project?

This is still a work in progress, and is not yet ready for production use.

## Is there a hosted version?

No. But there will be a Dockerfile you can use to run the Cache server.
