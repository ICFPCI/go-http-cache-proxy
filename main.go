package main

import (
	"caching-proxy/server/cache"
	"caching-proxy/server/handlers"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := flag.String("port", "", "The port on which the caching proxy server will run.")
	origin := flag.String("origin", "", "The URL of the server to which the requests will be forwarded.")
	clearCache := flag.Bool("clear-cache", false, "Clear the Redis cache when set.")
	flag.Parse()

	cache.InitRedisClient()
	if *clearCache {
		err := cache.ClearCache()
		if err != nil {
			log.Fatalf("Error clearing Redis cache: %v", err)
		} else {
			fmt.Println("Redis cache cleared successfully.")
		}
		return
	}

	if *port == "" && *origin == "" {
		fmt.Println("port and origin are needed to start the proxy")
		return
	} else {
		fmt.Println("this is my port:", *port)
		fmt.Println("this is my origin:", *origin)
	}

	os.Setenv("URL", *origin)

	router := http.NewServeMux()

	router.HandleFunc("/{path...}", handlers.HandleRequest)

	log.Printf("Server started at: http://localhost:%s", *port)
	http.ListenAndServe(fmt.Sprintf(":%s", *port), router)
}
