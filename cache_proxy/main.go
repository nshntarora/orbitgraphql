package main

import (
	"graphql_cache/cache_proxy/balancer"
	"graphql_cache/cache_proxy/cache_middleware"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const API_URL = "http://127.0.0.1:8080"

func main() {

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	e.Use(cache_middleware.CacheMiddleware)
	apiSever, err := url.Parse(API_URL)
	if err != nil {
		e.Logger.Fatal(err)
	}
	balancer := balancer.NewDefaultProxyBalancer(&middleware.ProxyTarget{
		URL: apiSever,
	})

	e.Use(middleware.Proxy(balancer))

	e.Logger.Fatal(e.Start(":9090"))
}
