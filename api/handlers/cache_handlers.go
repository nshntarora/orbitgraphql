package handlers

import (
	"bytes"
	"encoding/json"
	"graphql_cache/config"
	"graphql_cache/graphcache"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

func GetCacheHandler(cache *graphcache.GraphCache, cfg *config.Config) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		// Create a new HTTP request with the same method, URL, and body as the original request
		targetURL, err := url.Parse(cfg.Origin)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error parsing target URL", http.StatusInternalServerError)
		}

		proxyReq, err := http.NewRequest(r.Method, targetURL.String(), r.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		}

		// Copy the headers from the original request to the proxy request
		for name, values := range r.Header {
			for _, value := range values {
				proxyReq.Header.Add(name, value)
			}
		}

		requestBody, err := io.ReadAll(proxyReq.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, "error reading request body", http.StatusInternalServerError)
		}

		request := graphcache.GraphQLRequest{}
		request.FromBytes(requestBody)

		start := time.Now()

		astQuery, err := graphcache.GetASTFromQuery(request.Query)
		if err != nil {
			log.Println(err)
			http.Error(w, "error parsing query", http.StatusInternalServerError)
		}

		transformedBody, err := graphcache.AddTypenameToQuery(request.Query)
		if err != nil {
			log.Println(err)
			http.Error(w, "error transforming body", http.StatusInternalServerError)
		}

		log.Println("time taken to transform body ", time.Since(start))

		transformedRequest := request
		transformedRequest.Query = transformedBody

		if len(astQuery.Operations) > 0 && astQuery.Operations[0].Operation == "mutation" {
			// if the operation is a mutation, we don't cache it

			proxyReq.Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
			proxyReq.ContentLength = -1

			client := http.Client{}
			// Send the proxy request using the custom transport
			resp, err := client.Do(proxyReq)
			if err != nil || resp == nil {
				http.Error(w, "Error sending proxy request", http.StatusInternalServerError)

			}
			defer resp.Body.Close()

			// Copy the headers from the proxy response to the original response
			for name, values := range resp.Header {
				for _, value := range values {
					w.Header().Add(name, value)
				}
			}

			// Set the status code of the original response to the status code of the proxy response
			w.WriteHeader(resp.StatusCode)

			// Copy the body of the proxy response to the original response
			io.Copy(w, resp.Body)

			responseBody := new(bytes.Buffer)
			io.Copy(responseBody, resp.Body)

			responseMap := make(map[string]interface{})
			err = json.Unmarshal(responseBody.Bytes(), &responseMap)
			if err != nil {
				log.Println("Error unmarshalling response:", err, string(responseBody.Bytes()))
			}

			cache.InvalidateCache("data", responseMap, nil)
			return
		}

		cachedResponse, err := cache.ParseASTBuildResponse(astQuery, request)
		if err == nil && cachedResponse != nil {
			log.Println("serving response from cache...")
			br, err := json.Marshal(cachedResponse)
			if err != nil {
				http.Error(w, "error marshalling response", http.StatusInternalServerError)
			}
			log.Println("time taken to serve response from cache ", time.Since(start))
			graphqlresponse := graphcache.GraphQLResponse{Data: json.RawMessage(br)}
			res, err := cache.RemoveTypenameFromResponse(&graphqlresponse)
			if err != nil {
				http.Error(w, "error removing __typename", http.StatusInternalServerError)
			}
			w.Write(res.Bytes())
			return
		}

		proxyReq.Body = io.NopCloser(bytes.NewBuffer(transformedRequest.Bytes()))
		proxyReq.ContentLength = -1

		client := http.Client{}

		// Send the proxy request using the custom transport
		resp, err := client.Do(proxyReq)
		if err != nil {
			log.Println(err)
			http.Error(w, "error sending proxy request", http.StatusInternalServerError)
		}
		defer resp.Body.Close()

		// Copy the headers from the proxy response to the original response
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		responseBody := new(bytes.Buffer)
		io.Copy(responseBody, resp.Body)

		responseMap := make(map[string]interface{})
		err = json.Unmarshal(responseBody.Bytes(), &responseMap)
		if err != nil {
			log.Println("Error unmarshalling response:", err, string(responseBody.Bytes()))
		}

		log.Println("time taken to get response from API ", time.Since(start))

		astWithTypes, err := graphcache.GetASTFromQuery(transformedRequest.Query)
		if err != nil {
			log.Println(err)
			http.Error(w, "error parsing query", http.StatusInternalServerError)
		}

		log.Println("time taken to generate AST with types ", time.Since(start))

		reqVariables := transformedRequest.Variables
		variables := make(map[string]interface{})
		if reqVariables != nil {
			variables = reqVariables
		}

		for _, op := range astWithTypes.Operations {
			// for the operation op we need to traverse the response and build the relationship map where key is the requested field and value is the key where the actual response is stored in the cache
			responseKey := cache.GetQueryResponseKey(op, responseMap, variables)
			for key, value := range responseKey {
				if value != nil {
					cache.SetQueryCache(key, value)
				}
			}
		}

		log.Println("time taken to build response key ", time.Since(start))

		// go through the response. Every object that has a __typename field, and an id field cache it in the format of typename:id
		// for example, if the response has an object with __typename: "Organisation" and id: "1234", cache it as Organisation:1234
		// if the object has a nested object with __typename: "User" and id: "5678", cache
		// it as User:5678

		cache.CacheResponse("data", responseMap, nil)

		log.Println("time taken to cache response ", time.Since(start))

		// Copy the body of the proxy response to the original response
		// io.Copy(w, resp.Body)

		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, "error reading response body", http.StatusInternalServerError)
		}

		newResponse := &graphcache.GraphQLResponse{}
		newResponse.FromBytes(resBody)
		res, err := cache.RemoveTypenameFromResponse(newResponse)
		if err != nil {
			http.Error(w, "error removing __typename", http.StatusInternalServerError)
		}
		w.Write(res.Bytes())
		// w.WriteHeader(http.StatusOK)
	})
}
