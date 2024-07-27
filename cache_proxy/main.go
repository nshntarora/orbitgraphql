package main

import (
	"graphql_cache/cache_proxy/balancer"
	"graphql_cache/cache_proxy/cache_middleware"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const API_URL = "http://127.0.0.1:8080"

func main() {

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

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

	e.POST("/flush", func(c echo.Context) error {
		cache_middleware.Cache.Flush()
		return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
	})

	e.POST("/flushType", func(c echo.Context) error {
		flushByTypeRequest := FlushByTypeRequest{}
		err = c.Bind(&flushByTypeRequest)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"error": err.Error()})
		}
		cache_middleware.Cache.FlushByType(flushByTypeRequest.Type, flushByTypeRequest.ID)
		return c.JSON(http.StatusOK, map[string]interface{}{"success": true})
	})

	g := e.Group("")

	g.Use(cache_middleware.CacheMiddleware)

	g.Use(middleware.ProxyWithConfig(middleware.ProxyConfig{
		Balancer: balancer,
	}))

	e.Logger.Fatal(e.Start(":9090"))
}

type FlushByTypeRequest struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
