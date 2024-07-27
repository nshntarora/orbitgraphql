package main

import (
	"bytes"
	"fmt"
	"graphql_cache/cache_proxy/balancer"
	"graphql_cache/cache_proxy/cache_middleware"
	"graphql_cache/graphcache"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const API_URL = "http://127.0.0.1:8080"

func main() {

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	// e.Use(cache_middleware.NewCacheMiddleware(gc))
	apiSever, err := url.Parse(API_URL)
	if err != nil {
		e.Logger.Fatal(err)
	}
	balancer := balancer.NewDefaultProxyBalancer(&middleware.ProxyTarget{
		URL: apiSever,
	})

	e.GET("/debug", func(c echo.Context) error {
		cache_middleware.Cache.Debug()
		return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
	})

	g := e.Group("")

	g.Use(cache_middleware.CacheMiddleware)

	g.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
		ModifyResponse: func(resp *http.Response) error {
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				return err
			}
			newResponse := &graphcache.GraphQLResponse{}
			newResponse.FromBytes(responseBody)
			res, err := cache_middleware.Cache.RemoveTypenameFromResponse(newResponse)
			if err != nil {
				fmt.Println("Error removing __typename:", err)
				return nil
			}
			body := io.NopCloser(bytes.NewReader(res.Bytes()))
			resp.Body = body
			resp.ContentLength = int64(len(res.Bytes()))
			resp.Header.Set("Content-Length", strconv.Itoa(len(res.Bytes())))
			return nil
		},
	}))

	e.Logger.Fatal(e.Start(":9090"))
}
