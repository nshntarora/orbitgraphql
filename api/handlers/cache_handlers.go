package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"io"
	"net/http"
	"net/url"
	"time"
)

const CACHE_STATUS_BYPASS = "BYPASS"
const CACHE_STATUS_HIT = "HIT"
const CACHE_STATUS_MISS = "MISS"

func GetCacheHandler(cache *graphcache.GraphCache, cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Add("Content-Type", "application/json")

		// Create a new HTTP request with the same method, URL, and body as the original request
		targetURL, err := url.Parse(cfg.Origin)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
		}

		// only handle if the request is of content type application/json
		// for all other content types, pass the request to the origin server
		if r.Header.Get("Content-Type") != "application/json" {
			proxyReq, err := CopyRequest(r, cfg.Origin)
			if err != nil {
				fmt.Println(err)
				http.Error(w, "Error copying request", http.StatusInternalServerError)
			}

			resp, err := SendRequest(proxyReq, w, map[string]interface{}{
				cfg.CacheHeaderName: CACHE_STATUS_BYPASS,
			})
			if err != nil {
				fmt.Println(err)
				http.Error(w, "error sending proxy request", http.StatusInternalServerError)
			}
			defer resp.Body.Close()

			responseBody := new(bytes.Buffer)
			io.Copy(responseBody, resp.Body)
			w.Write(responseBody.Bytes())
			return
		}

		proxyReq, err := CopyRequest(r, targetURL.String())
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Error copying request", http.StatusInternalServerError)
		}

		requestBody, err := io.ReadAll(proxyReq.Body)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "error reading request body", http.StatusInternalServerError)
		}

		request := graphcache.GraphQLRequest{}
		request.FromBytes(requestBody)

		varStr := ""
		for key, value := range request.Variables {
			varStr = varStr + key + ":" + fmt.Sprintf("%v", value)
		}

		cacheKeyPrefix := base64.StdEncoding.EncodeToString([]byte(request.Query + varStr))

		start := time.Now()

		astQuery, err := graphcache.GetASTFromQuery(request.Query)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "error parsing query", http.StatusInternalServerError)
		}

		transformedBody, err := graphcache.AddTypenameToQuery(request.Query)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "error transforming body", http.StatusInternalServerError)
		}

		fmt.Println("time taken to transform body ", time.Since(start))

		transformedRequest := request
		transformedRequest.Query = transformedBody

		if len(astQuery.Operations) > 0 && astQuery.Operations[0].Operation == "mutation" {
			// if the operation is a mutation, we don't cache it

			proxyReq.Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
			proxyReq.ContentLength = -1

			resp, err := SendRequest(proxyReq, w, map[string]interface{}{
				cfg.CacheHeaderName: CACHE_STATUS_BYPASS,
			})
			if err != nil {
				fmt.Println(err)
				http.Error(w, "error sending proxy request", http.StatusInternalServerError)
			}
			defer resp.Body.Close()

			responseBody := new(bytes.Buffer)
			io.Copy(responseBody, resp.Body)

			responseMap := make(map[string]interface{})
			err = json.Unmarshal(responseBody.Bytes(), &responseMap)
			if err != nil {
				fmt.Println("Error unmarshalling response:", string(responseBody.Bytes()))
			}

			cache.InvalidateCache("data", responseMap, nil)

			newResponse := &graphcache.GraphQLResponse{}
			newResponse.FromBytes(responseBody.Bytes())
			res, err := cache.RemoveTypenameFromResponse(newResponse)
			if err != nil {
				http.Error(w, "error removing __typename", http.StatusInternalServerError)
			}

			w.Write(res.Bytes())
			return
		}

		cachedResponse, err := cache.ParseASTBuildResponse(cacheKeyPrefix, astQuery, request)
		if err == nil && cachedResponse != nil {
			fmt.Println("serving response from cache...")
			br, err := json.Marshal(cachedResponse)
			if err != nil {
				http.Error(w, "error marshalling response", http.StatusInternalServerError)
			}
			fmt.Println("time taken to serve response from cache ", time.Since(start))
			graphqlresponse := graphcache.GraphQLResponse{Data: json.RawMessage(br)}
			res, err := cache.RemoveTypenameFromResponse(&graphqlresponse)
			if err != nil {
				http.Error(w, "error removing __typename", http.StatusInternalServerError)
			}
			w.Header().Add(cfg.CacheHeaderName, CACHE_STATUS_HIT)
			w.Write(res.Bytes())
			return
		}

		proxyReq.Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
		proxyReq.ContentLength = -1

		resp, err := SendRequest(proxyReq, w, map[string]interface{}{
			cfg.CacheHeaderName: CACHE_STATUS_MISS,
		})
		if err != nil {
			fmt.Println(err)
			http.Error(w, "error sending proxy request", http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		responseBody := new(bytes.Buffer)
		io.Copy(responseBody, resp.Body)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(responseBody.Bytes(), &responseMap)
		if err != nil {
			fmt.Println("Error unmarshalling response:", string(responseBody.Bytes()))
		}

		fmt.Println("time taken to get response from API ", time.Since(start))

		astWithTypes, err := graphcache.GetASTFromQuery(transformedRequest.Query)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "error parsing query", http.StatusInternalServerError)
		}

		fmt.Println("time taken to generate AST with types ", time.Since(start))

		reqVariables := transformedRequest.Variables
		variables := make(map[string]interface{})
		if reqVariables != nil {
			variables = reqVariables
		}

		for _, op := range astWithTypes.Operations {
			// for the operation op we need to traverse the response and build the relationship map where key is the requested field and value is the key where the actual response is stored in the cache
			responseKey := cache.GetQueryResponseKey(cacheKeyPrefix, op, responseMap, variables)
			for key, value := range responseKey {
				if value != nil {
					cache.SetQueryCache(key, value)
				}
			}
		}

		fmt.Println("time taken to build response key ", time.Since(start))

		// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
		// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
		// if the object has a nested object with __typename: "User" and id: "5678", cache
		// it as User:5678

		cache.CacheResponse(cacheKeyPrefix, "data", responseMap, nil)

		fmt.Println("time taken to cache response ", time.Since(start), responseMap)

		newResponse := &graphcache.GraphQLResponse{}
		newResponse.FromBytes(responseBody.Bytes())
		res, err := cache.RemoveTypenameFromResponse(newResponse)
		if err != nil {
			http.Error(w, "error removing __typename", http.StatusInternalServerError)
		}

		w.Write(res.Bytes())
	})
}

func CopyRequest(r *http.Request, targetURL string) (*http.Request, error) {
	proxyReq, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
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

func SendRequest(proxyReq *http.Request, w http.ResponseWriter, headers map[string]interface{}) (*http.Response, error) {
	client := http.Client{}
	// Send the proxy request using the custom transport
	resp, err := client.Do(proxyReq)
	if err != nil || resp == nil {
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
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

	return resp, nil
}
