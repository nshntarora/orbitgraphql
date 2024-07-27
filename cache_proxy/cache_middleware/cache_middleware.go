package cache_middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"graphql_cache/graphcache"
	"graphql_cache/transformer"
	"graphql_cache/utils/ast_utils"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

var Cache = graphcache.NewGraphCache("redis")

func CacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		requestBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			fmt.Println("Error reading request body:", err)
			return nil
		}

		fmt.Println("request content length ", c.Request().ContentLength)
		fmt.Println("request body length", len(requestBytes))

		var request graphcache.GraphQLRequest
		err = json.Unmarshal(requestBytes, &request)
		if err != nil {
			fmt.Println("Error unmarshalling request:", err)
			return nil
		}

		astQuery, err := ast_utils.GetASTFromQuery(request.Query)
		if err != nil {
			fmt.Println("Error parsing query:", err)
			return nil
		}

		if astQuery.Operations[0].Operation == "mutation" {
			// if the operation is a mutation, we don't cache it
			// c.Request().Body = io.NopCloser(bytes.NewBuffer(requestBytes))
			// c.Request().ContentLength = -1
			transformedBody, err := transformer.TransformBody(request.Query, astQuery)
			if err != nil {
				fmt.Println("Error transforming body:", err)
				return nil
			}

			fmt.Println("time taken to transform body ", time.Since(start))

			transformedRequest := request
			transformedRequest.Query = transformedBody

			c.Request().Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
			c.Request().ContentLength = -1

			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
			c.Response().Writer = writer
			err = next(c)
			if err != nil {
				return err
			}
			fmt.Println("found a mutation, invalidating cache...")

			responseMap := make(map[string]interface{})
			err = json.Unmarshal(resBody.Bytes(), &responseMap)
			if err != nil {
				fmt.Println("Error unmarshalling response:", err)
			}

			// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
			// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
			// if the object has a nested object with __typename: "User" and id: "5678", cache
			// it as User:5678

			Cache.InvalidateCache("data", responseMap, nil)

			// newResponse := &graphcache.GraphQLResponse{}
			// newResponse.FromBytes(resBody.Bytes())
			// res, err := Cache.RemoveTypenameFromResponse(newResponse)
			// if err != nil {
			// 	fmt.Println("Error removing __typename:", err)
			// 	return nil
			// }
			// fmt.Println("response bytes ", string(res.Bytes()))
			// c.Response().Write(res.Bytes())
			// c.Response().Header().Set("X-Proxy", "GraphQL Cache")
			return nil
		}

		cachedResponse, err := Cache.ParseASTBuildResponse(astQuery, request)
		if err == nil && cachedResponse != nil {
			fmt.Println("serving response from cache...")
			br, err := json.Marshal(cachedResponse)
			if err != nil {
				return err
			}
			fmt.Println("time taken to serve response from cache ", time.Since(start))
			graphqlresponse := graphcache.GraphQLResponse{Data: json.RawMessage(br)}
			res, err := Cache.RemoveTypenameFromResponse(&graphqlresponse)
			if err != nil {
				fmt.Println("Error removing __typename:", err)
				return nil
			}
			return c.JSON(200, res)
		}

		transformedBody, err := transformer.TransformBody(request.Query, astQuery)
		if err != nil {
			fmt.Println("Error transforming body:", err)
			return nil
		}

		fmt.Println("time taken to transform body ", time.Since(start))

		transformedRequest := request
		transformedRequest.Query = transformedBody

		c.Request().Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
		c.Request().ContentLength = -1

		resBody := new(bytes.Buffer)
		mw := io.MultiWriter(c.Response().Writer, resBody)
		writer := &bodyDumpResponseWriter{Writer: mw, ResponseWriter: c.Response().Writer}
		c.Response().Writer = writer
		err = next(c)
		if err != nil {
			return err
		}
		responseMap := make(map[string]interface{})
		err = json.Unmarshal(resBody.Bytes(), &responseMap)
		if err != nil {
			fmt.Println("Error unmarshalling response:", err)
		}

		fmt.Println("time taken to get response from API ", time.Since(start))

		astWithTypes, err := ast_utils.GetASTFromQuery(transformedRequest.Query)
		if err != nil {
			fmt.Println("Error parsing query:", err)
			return nil
		}

		fmt.Println("time taken to generate AST with types ", time.Since(start))

		reqVariables := transformedRequest.Variables
		variables := make(map[string]interface{})
		if reqVariables != nil {
			variables = reqVariables
		}

		for _, op := range astWithTypes.Operations {
			// for the operation op we need to traverse the response and the ast together to build a graph of the relations

			// build the relation graph
			responseKey := Cache.GetQueryResponseKey(op, responseMap, variables)
			for key, value := range responseKey {
				if value != nil {
					Cache.SetQueryCache(key, value)
				}
			}
		}

		fmt.Println("time taken to build response key ", time.Since(start))

		// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
		// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
		// if the object has a nested object with __typename: "User" and id: "5678", cache
		// it as User:5678

		Cache.CacheResponse("data", responseMap, nil)

		fmt.Println("time taken to cache response ", time.Since(start))

		// cacheStore.Debug("cacheStore")
		// recordCacheStore.Debug("recordCacheStore")
		// queryCacheStore.Debug("queryCacheStore")

		// cacheState, _ := cacheStore.JSON()
		// fmt.Println(string(cacheState))

		// recordCacheState, _ := recordCacheStore.JSON()
		// fmt.Println(string(recordCacheState))

		// queryCacheState, _ := queryCacheStore.JSON()
		// fmt.Println(string(queryCacheState))

		fmt.Println("time taken to finish completely ", time.Since(start))
		newResponse := &graphcache.GraphQLResponse{}
		newResponse.FromBytes(resBody.Bytes())
		Cache.RemoveTypenameFromResponse(newResponse)
		c.Response().Header().Set("X-Proxy", "GraphQL Cache")
		return nil
	}
}

type bodyDumpResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *bodyDumpResponseWriter) WriteHeader(code int) {
	w.ResponseWriter.WriteHeader(code)
}

func (w *bodyDumpResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *bodyDumpResponseWriter) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *bodyDumpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}
