package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/joho/godotenv"
)

type service struct {
	cache      *lru.Cache
	inProgress *sync.Map // stores the set of IDs being processed
	timeout    time.Duration
}

type resizeRequest struct {
	URLs   []string `json:"urls"`
	Width  uint     `json:"width"`
	Height uint     `json:"height"`
}

type resizeResult struct {
	Result string `json:"result"`
	URL    string `json:"url,omitempty"`
	OldURL string `json:"original_url"`
	Cached bool   `json:"cached"`
}

const (
	proto      = "http://"
	hostport   = "localhost:8080"
	success    = "success"
	failure    = "failure"
	processing = "processing"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using defaults")
	}

	// Get timeout from env or use default
	timeoutStr := os.Getenv("IMAGE_PROCESSING_TIMEOUT")
	timeout := 30 * time.Second // default timeout
	if timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = time.Duration(t) * time.Second
		}
	}

	cache, err := lru.New(1024)
	if err != nil {
		log.Panicf("Failed to create cache: %v", err)
	}

	svc := &service{
		cache:      cache,
		inProgress: &sync.Map{},
		timeout:    timeout,
	}

	mux := http.NewServeMux()
	mux.Handle("/v1/resize", svc.resizeHandler())
	mux.Handle("/v1/image/", svc.getImageHandler())
	address := hostport

	log.Print("Listening on ", hostport)
	// When running on docker mac, can't listen only on localhost
	panic(http.ListenAndServe(address, mux))
}
