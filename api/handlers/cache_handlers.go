package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"graphql_cache/logger"
	"io"
	"net/http"
	"net/url"
	"time"
)

const CACHE_STATUS_BYPASS = "BYPASS"
const CACHE_STATUS_HIT = "HIT"
const CACHE_STATUS_MISS = "MISS"

func CreateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func GetCacheHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestId := CreateRequestID()
		startTime := time.Now()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		ctx = logger.SetMetadata(ctx, map[string]interface{}{
			"request_id":     requestId,
			"method":         r.Method,
			"remote_address": r.RemoteAddr,
			"host":           r.Host,
			"type":           "request",
			"content_type":   r.Header.Get("Content-Type"),
			"time_started":   startTime.Format(time.RFC3339),
			"path":           r.URL.Path,
			"user_agent":     r.Header.Get("User-Agent"),
		})
		ctx = CacheMiddleware(ctx, cfg, w, r)
		ctx = logger.SetMetadata(ctx, map[string]interface{}{
			"status":         ctx.Value("status"),
			"content_length": ctx.Value("contentLength"),
			"time_taken":     time.Since(startTime).String(),
			"operation_name": ctx.Value("operationName"),
			"cache_status":   w.Header().Get(cfg.CacheHeaderName),
		})
		logger.Info(ctx)
	}
}

func CacheMiddleware(ctx context.Context, cfg *config.Config, w http.ResponseWriter, r *http.Request) context.Context {
	w.Header().Add("Content-Type", "application/json")
	// Create a new HTTP request with the same method, URL, and body as the original request
	targetURL, err := url.Parse(cfg.Origin)
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// only handle if the request is of content type application/json
	// for all other content types, pass the request to the origin server
	if r.Header.Get("Content-Type") != "application/json" {
		proxyReq, err := CopyRequest(ctx, r, cfg.Origin)
		if err != nil {
			logger.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		resp, err := SendRequest(&ctx, proxyReq, w, map[string]interface{}{
			cfg.CacheHeaderName: CACHE_STATUS_BYPASS,
		})
		if err != nil {
			logger.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		responseBody := new(bytes.Buffer)
		io.Copy(responseBody, resp.Body)
		w.Write(responseBody.Bytes())
		return ctx
	}

	proxyReq, err := CopyRequest(ctx, r, targetURL.String())
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	requestBody, err := io.ReadAll(proxyReq.Body)
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	proxyReq.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	request := graphcache.GraphQLRequest{}
	request.FromBytes(requestBody)
	ctx = context.WithValue(ctx, "operationName", request.OperationName)

	cache := graphcache.NewGraphCacheWithOptions(ctx, GetCacheOptions(cfg, GetScopeValues(cfg, proxyReq)))

	start := time.Now()

	astQuery, err := graphcache.GetASTFromQuery(request.Query)
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	transformedBody, err := graphcache.AddTypenameToQuery(request.Query)
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	logger.Debug(ctx, "time taken to transform body ", time.Since(start))

	transformedRequest := request
	transformedRequest.Query = transformedBody

	if len(astQuery.Operations) > 0 && astQuery.Operations[0].Operation == "mutation" {
		// if the operation is a mutation, we don't cache it

		proxyReq.Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
		proxyReq.ContentLength = -1

		resp, err := SendRequest(&ctx, proxyReq, w, map[string]interface{}{
			cfg.CacheHeaderName: CACHE_STATUS_BYPASS,
		})
		if err != nil {
			logger.Error(ctx, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		responseBody := new(bytes.Buffer)
		io.Copy(responseBody, resp.Body)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(responseBody.Bytes(), &responseMap)
		if err != nil {
			logger.Error(ctx, err)
		}

		cache.InvalidateCache("data", responseMap, nil)

		newResponse := &graphcache.GraphQLResponse{}
		newResponse.FromBytes(responseBody.Bytes())
		res, err := cache.RemoveTypenameFromResponse(newResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Write(res.Bytes())
		return ctx
	}

	cachedResponse, err := cache.ParseASTBuildResponse(astQuery, request)
	if err == nil && cachedResponse != nil {
		logger.Debug(ctx, "serving response from cache")
		br, err := json.Marshal(cachedResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		logger.Debug(ctx, "time taken to serve response from cache ", time.Since(start))
		graphqlresponse := graphcache.GraphQLResponse{Data: json.RawMessage(br)}
		res, err := cache.RemoveTypenameFromResponse(&graphqlresponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Add(cfg.CacheHeaderName, CACHE_STATUS_HIT)
		w.Write(res.Bytes())
		ctx = context.WithValue(ctx, "status", http.StatusOK)
		ctx = context.WithValue(ctx, "contentLength", len(res.Bytes()))
		return ctx
	}

	proxyReq.Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
	proxyReq.ContentLength = -1

	resp, err := SendRequest(&ctx, proxyReq, w, map[string]interface{}{
		cfg.CacheHeaderName: CACHE_STATUS_MISS,
	})
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	responseBody := new(bytes.Buffer)
	io.Copy(responseBody, resp.Body)

	responseMap := make(map[string]interface{})
	err = json.Unmarshal(responseBody.Bytes(), &responseMap)
	if err != nil {
		logger.Error(ctx, err)
	}

	logger.Debug(ctx, "time taken to get response from API ", time.Since(start))

	astWithTypes, err := graphcache.GetASTFromQuery(transformedRequest.Query)
	if err != nil {
		logger.Error(ctx, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	logger.Debug(ctx, "time taken to generate AST with types ", time.Since(start))

	reqVariables := transformedRequest.Variables
	variables := make(map[string]interface{})
	if reqVariables != nil {
		variables = reqVariables
	}

	for _, op := range astWithTypes.Operations {
		// for the operation op we need to traverse the response and build the relationship map where key is the requested field and value is the key where the actual response is stored in the cache
		cache.CacheOperation(op, responseMap, variables)
	}

	logger.Debug(ctx, "time taken to build response key ", time.Since(start))

	// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
	// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
	// if the object has a nested object with __typename: "User" and id: "5678", cache
	// it as User:5678
	cache.CacheResponse("data", responseMap, nil)

	logger.Debug(ctx, "time taken to cache response ", time.Since(start), responseMap)

	newResponse := &graphcache.GraphQLResponse{}
	newResponse.FromBytes(responseBody.Bytes())
	res, err := cache.RemoveTypenameFromResponse(newResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	w.Write(res.Bytes())
	return ctx
}

func CopyRequest(ctx context.Context, r *http.Request, targetURL string) (*http.Request, error) {
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	// Copy the headers from the original request to the proxy request
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	proxyReq.ContentLength = -1

	return proxyReq, nil
}

func SendRequest(ctx *context.Context, proxyReq *http.Request, w http.ResponseWriter, headers map[string]interface{}) (*http.Response, error) {
	client := http.Client{}
	// Send the proxy request using the custom transport
	resp, err := client.Do(proxyReq)
	if err != nil || resp == nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return resp, err
	}

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		if name != "Content-Length" {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}
	}

	for key, value := range headers {
		w.Header().Add(key, fmt.Sprintf("%v", value))
	}

	w.WriteHeader(resp.StatusCode)
	*ctx = context.WithValue(*ctx, "status", resp.StatusCode)
	*ctx = context.WithValue(*ctx, "contentLength", resp.ContentLength)

	return resp, nil
}
