---
icon: question
---

# How does it work?

The cache server caches based on the `__typename` and a primary key in your response objects. `__typename` is an internal field in every GraphQL type, and I assume every unique object in your GraphQL will have a primary key that you will sending to your frontend.

When a request hits the Orbit GraphQL server, it converts it into an AST (abstract syntax tree), then appends the `__typename` field to every parent field and the query.

Using the `__typename` and `id` (_primary key - configurable_) fields in your response it builds a cache for the returned objects.

On subsequent requests, if the same `query` is made again with the same `variables`, and the result exists in the cache, Orbit builds the response object itself and sends that to your client.

For every `mutation` that hits the Orbit server, it forwards the request to the origin to make the mutation, and then checks the `__typename` and `id` fields returned by the mutation. Based on the response that is received, we know which object was updated and use it to invalidate the cache accordingly.

You can also invalidate the cache manually using the cache purging APIs.

This is not production ready yet.

Here is a non-exhaustive list of things planned for the project:

1. Support for Fragments.
2. Benchmarking.
3. Go/JavaScript clients for the administration APIs (used to flush cache).
4. Better observability setup (to help monitor how the cache server is performing).
5. Support for analytics on top of your GraphQL API to help you get insights on how your API is being consumed.

Currently, there's no plan to offer a hosted version for this (I'm scratching my own itch but, never say never)
