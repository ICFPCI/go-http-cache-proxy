package server

import (
	"caching-proxy/server/cache"
	"caching-proxy/server/handlers"
	"fmt"
	"log"
	"net/http"
	"os"
)

func StartServer(port *string, url *string) {
	router := http.NewServeMux()

	os.Setenv("URL", *url)

	cache.InitRedisClient()

	router.HandleFunc("/{path...}", handlers.HandleRequest)

	log.Printf("Server started at: http://localhost:%s", *port)
	http.ListenAndServe(fmt.Sprintf(":%s", *port), router)
}
