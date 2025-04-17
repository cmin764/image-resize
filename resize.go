package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	jpgresize "github.com/nfnt/resize"
)

func (s *service) processResizes(request resizeRequest, async bool) ([]resizeResult, error) {
	mode := "sync"
	if async {
		mode = "async"
	}
	log.Printf("Processing %d images in %s mode\n", len(request.URLs), mode)

	results := make([]resizeResult, 0, len(request.URLs))
	resultChan := make(chan resizeResult, len(request.URLs))
	var wg sync.WaitGroup

	for _, url := range request.URLs {
		result := resizeResult{}
		id := genID(url)
		result.OldURL = url
		result.Cached = false
		result.URL = proto + hostport + "/v1/image/" + id + ".jpeg"

		// Check first if the image is already cached.
		if s.cache.Contains(id) {
			result.Result = success
			result.Cached = true
			results = append(results, result)
			continue
		}

		// Otherwise, check if the image is already being processed.
		if _, progress := s.inProgress.Load(id); progress {
			result.Result = processing
			results = append(results, result)
			continue
		}

		// If none, we should process it and mark the status accordingly.
		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			resizeCh := make(chan struct{})
			result.Result = processing
			go func() {
				s.inProgress.Store(id, struct{}{})
				defer s.inProgress.Delete(id) // image processing would be done by the end

				data, err := fetchAndResize(url, request.Width, request.Height)
				if err != nil {
					log.Printf("failed to resize %s: %v", url, err)
					result.Result = failure
					result.URL = ""
				} else {
					log.Printf("succeeded to resize %s", url)
					result.Result = success
					result.Cached = false
					log.Print("adding image to cache ", id)
					s.cache.Add(id, data)
				}
				resizeCh <- struct{}{}
			}()
			if !async {
				<-resizeCh
			}

			// log.Println("Adding result to channel", result)
			resultChan <- result
			// log.Println("Added result to channel", result)
		}(url)
	}

	// Close channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	log.Println("Waiting for all the images to complete...")
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

func fetchAndResize(url string, width uint, height uint) ([]byte, error) {
	data, err := fetch(url)
	if err != nil {
		return nil, err
	}

	return resize(data, width, height)
}

func fetch(url string) ([]byte, error) {
	log.Print("fetching ", url)
	r, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %v", err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 status: %d", r.StatusCode)
	}

	data, err := ioutil.ReadAll(io.LimitReader(r.Body, 15*1024*1024))
	if err != nil {
		return nil, fmt.Errorf("failed to read fetch data: %v", err)
	}

	return data, nil
}

func resize(data []byte, width uint, height uint) ([]byte, error) {
	// decode jpeg into image.Image
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to jped decode: %v", err)
	}

	var newImage image.Image

	// if either width or height is 0, it will resize respecting the aspect ratio
	newImage = jpgresize.Resize(width, height, img, jpgresize.Lanczos3)

	newData := bytes.Buffer{}
	err = jpeg.Encode(bufio.NewWriter(&newData), newImage, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to jpeg encode resized image: %v", err)
	}

	return newData.Bytes(), nil
}

func genID(url string) string {
	hash := sha256.Sum256([]byte(url))
	return base64.URLEncoding.EncodeToString(hash[:])
}
