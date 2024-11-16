package handlers

import (
	"caching-proxy/server/cache"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheStatus string

const (
	CacheHit  CacheStatus = "HIT"
	CacheMiss CacheStatus = "MISS"
)

func getDataFromURL(path string) ([]byte, int, error) {
	apiURL := fmt.Sprintf("%s/%s", os.Getenv("URL"), path)

	log.Printf("Fetching data from: %s", apiURL)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to fetch data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("failed to read response: %w", err)
	}

	return body, resp.StatusCode, nil
}

func setCacheHeaders(w http.ResponseWriter, status CacheStatus) {
	w.Header().Set("X-Cache", string(status))
}

func verifyCache(ctx context.Context, path string) (bool, []byte, error) {
	cmd := cache.RedisClient.Get(ctx, path)

	if cmd.Err() == redis.Nil {
		log.Println(cmd.Err().Error())
		return false, nil, nil
	}
	if cmd.Err() != nil {
		return false, nil, fmt.Errorf("cache error: %w", cmd.Err())
	}

	data := []byte(cmd.Val())
	return true, data, nil
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	path := r.PathValue("path")
	if path == "" {
		path = "/"
	}

	isInCache, cachedData, err := verifyCache(r.Context(), path)
	if err != nil {
		log.Printf("Cache verification error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if isInCache {
		log.Print("Cache hit")
		setCacheHeaders(w, CacheHit)
		w.Write(cachedData)
		return
	}

	data, statusCode, err := getDataFromURL(path)
	if err != nil {
		log.Printf("Upstream fetch error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	setCacheHeaders(w, CacheMiss)
	w.WriteHeader(statusCode)

	if statusCode == http.StatusOK {
		if err := cache.RedisClient.Set(r.Context(), path, data, 12*time.Hour).Err(); err != nil {
			log.Printf("Cache storage error: %v", err)
		}
	}

	w.Write(data)
}
