---
description: A cache server for your GraphQL API so you're ready for unexpected load
icon: brackets-round
cover: .gitbook/assets/Frame 13.png
coverY: 0
layout:
  cover:
    visible: true
    size: full
  title:
    visible: true
  description:
    visible: true
  tableOfContents:
    visible: true
  outline:
    visible: true
  pagination:
    visible: true
---

# Orbit GraphQL

**What?**

1. Deploy it in front of GraphQL API, and it will start caching all requests passing through it.
2. Your cache gets automatically invalidated if anything changes in your application
3. Queries are only sent to origin if there is a cache MISS

**Wait, why is it needed?**

Because all GraphQL requests are on a single endpoint (usually `POST`) and resources are differentiated based on what your request body looks like, we can't use HTTP caching methods for them (say goodbye to etags, 304s, cache-control etc.)

Currently the ecosystem solves for the problem in two ways, client side cache or a server side cache.

Clients like [urql](https://github.com/urql-graphql/urql) can cache your API responses on the client side. Services like [Stellate](https://stellate.co/) act as a cache proxy that does the same thing on the server side.

Orbit GraphQL is an open-source alternative to a tool like Stellate (server side cache for your GraphQL API).

{% hint style="warning" %}
This project is not ready for production traffic yet, but hopefully it will be soon (if you [help](https://github.com/nshntarora/orbitgraphql))
{% endhint %}
