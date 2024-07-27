package cache_middleware

import (
	"bufio"
	"bytes"
	"encoding/json"
	"graphql_cache/graphcache"
	"io"
	"net"
	"net/http"
	"time"

	"log"

	"github.com/labstack/echo/v4"
)

var Cache = graphcache.NewGraphCache("redis")

// CacheMiddleware is a middleware that intercepts the request and response
// and caches the response if it is a query
// if it is a mutation, it does not cache the response, but invalidates the objects which have been mutated
func CacheMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		requestBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			log.Println("Error reading request body:", err)
			return nil
		}

		var request graphcache.GraphQLRequest
		err = json.Unmarshal(requestBytes, &request)
		if err != nil {
			log.Println("Error unmarshalling request:", err)
			return nil
		}

		astQuery, err := graphcache.GetASTFromQuery(request.Query)
		if err != nil {
			log.Println("Error parsing query:", err)
			return nil
		}

		if astQuery.Operations[0].Operation == "mutation" {
			// if the operation is a mutation, we don't cache it
			transformedBody, err := graphcache.AddTypenameToQuery(request.Query)
			if err != nil {
				log.Println("Error transforming body:", err)
				return nil
			}

			log.Println("time taken to transform body ", time.Since(start))

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
				log.Println("Error unmarshalling response:", err)
			}

			Cache.InvalidateCache("data", responseMap, nil)
			return nil
		}

		cachedResponse, err := Cache.ParseASTBuildResponse(astQuery, request)
		if err == nil && cachedResponse != nil {
			log.Println("serving response from cache...")
			br, err := json.Marshal(cachedResponse)
			if err != nil {
				return err
			}
			log.Println("time taken to serve response from cache ", time.Since(start))
			graphqlresponse := graphcache.GraphQLResponse{Data: json.RawMessage(br)}
			res, err := Cache.RemoveTypenameFromResponse(&graphqlresponse)
			if err != nil {
				log.Println("Error removing __typename:", err)
				return nil
			}
			return c.JSON(200, res)
		}

		transformedBody, err := graphcache.AddTypenameToQuery(request.Query)
		if err != nil {
			log.Println("Error transforming body:", err)
			return nil
		}

		log.Println("time taken to transform body ", time.Since(start))

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
			log.Println("Error unmarshalling response:", err)
		}

		log.Println("time taken to get response from API ", time.Since(start))

		astWithTypes, err := graphcache.GetASTFromQuery(transformedRequest.Query)
		if err != nil {
			log.Println("Error parsing query:", err)
			return nil
		}

		log.Println("time taken to generate AST with types ", time.Since(start))

		reqVariables := transformedRequest.Variables
		variables := make(map[string]interface{})
		if reqVariables != nil {
			variables = reqVariables
		}

		for _, op := range astWithTypes.Operations {
			// for the operation op we need to traverse the response and build the relationship map where key is the requested field and value is the key where the actual response is stored in the cache
			responseKey := Cache.GetQueryResponseKey(op, responseMap, variables)
			for key, value := range responseKey {
				if value != nil {
					Cache.SetQueryCache(key, value)
				}
			}
		}

		log.Println("time taken to build response key ", time.Since(start))

		// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
		// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
		// if the object has a nested object with __typename: "User" and id: "5678", cache
		// it as User:5678

		Cache.CacheResponse("data", responseMap, nil)

		log.Println("time taken to cache response ", time.Since(start))

		// remove __typename from the response
		newResponse := &graphcache.GraphQLResponse{}
		newResponse.FromBytes(resBody.Bytes())
		Cache.RemoveTypenameFromResponse(newResponse)
		c.Response().Header().Set("X-Proxy", "GraphQL Cache")
		return nil
	}
}

// bodyDumpResponseWrite is a custom response writer that writes the response body to a buffer
// and also writes it to the actual response writer
// this is used to read the response body after it has been written to the response writer
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
