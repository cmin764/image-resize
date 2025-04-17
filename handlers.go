package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func (s *service) resizeHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("Expecting POST request"))
			return
		}

		request := resizeRequest{}
		err := json.NewDecoder(io.LimitReader(r.Body, 8*1024)).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to parse request"))
			return
		}

		async := strings.ToLower(r.URL.Query().Get("async")) == "true"
		results, err := s.processResizes(request, async)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to process request"))
			return
		}

		data, err := json.Marshal(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to marshal response"))
			return
		}

		// When all the requested images are cached, we're not creating any new resources.
		allCached := true
		for _, result := range results {
			if !result.Cached {
				allCached = false
				break
			}
		}

		if async {
			w.WriteHeader(http.StatusAccepted) // for later on processing
		} else if allCached {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
		w.Header().Add("content-type", "application/json")
		w.Write(data)
	})
}

func (s *service) getImageHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Print("fetching ", r.URL.String())
		// Extract ID by splitting path and removing extension
		path := strings.TrimRight(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		filename := parts[len(parts)-1]
		id := strings.Split(filename, ".")[0]

		// Check if image is being processed
		if _, processing := s.inProgress.Load(id); processing {
			// Wait until processing is complete or timeout occurs
			done := make(chan struct{})
			go func() {
				for {
					if _, stillProcessing := s.inProgress.Load(id); !stillProcessing {
						close(done)
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
			}()

			// Wait with timeout
			select {
			case <-done:
				// Processing finished, continue to check cache
			case <-time.After(s.timeout):
				w.WriteHeader(http.StatusProcessing) // 102 Processing
				w.Header().Add("content-type", "application/json")
				// Do not mind if the response below isn't flushed/displayed.
				resp := map[string]string{
					"status":  "processing",
					"message": "Image is still being processed, please try again later",
				}
				if err := json.NewEncoder(w).Encode(resp); err != nil {
					log.Printf("Failed to encode response: %v", err)
				}
				return
			}
		}

		// Check cache after processing is complete
		data, ok := s.cache.Get(id)
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Add("content-type", "image/jpeg")
		w.Write(data.([]byte))
	})
}
