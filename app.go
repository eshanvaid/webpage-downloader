package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

type Request struct {
	URL        string `json:"url"`
	RetryLimit int    `json:"retryLimit"`
}

type Response struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	SourceURL string `json:"sourceUrl"`
}

type CacheItem struct {
	Body      []byte
	Timestamp time.Time
	ID        string
}

var cache map[string]CacheItem

const cacheExpiryInterval = 24 * time.Hour

var (
	numActiveWorkers = 10
	semaphore        = make(chan struct{}, numActiveWorkers)
	wg               sync.WaitGroup
)

func downloadPageSource(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request.", http.StatusBadRequest)
		return
	}

	// Generate a unique ID and filename for the webpage based on the URL
	hasher := sha1.New()
	hasher.Write([]byte(req.URL))
	id := hex.EncodeToString(hasher.Sum(nil))
	filename := fmt.Sprintf("files/%s.html", id)

	// Check for the directory's existence and create it if it doesn't exist
	if _, err := os.Stat("files"); os.IsNotExist(err) {
		os.Mkdir("files", os.ModePerm)
	}

	// Generate the response object
	res := Response{
		ID:        id,
		URL:       req.URL,
		SourceURL: filename,
	}
	jsonStr, _ := json.MarshalIndent(res, "", "  ")

	// Check if the webpage is in the cache
	if item, ok := cache[req.URL]; ok {
		// Check if the webpage was requested within the last 24 hours
		if time.Since(item.Timestamp) < cacheExpiryInterval {
			// Create a file with the unique ID as the filename
			file, err := os.Create(filename)
			if err != nil {
				http.Error(w, "Error creating file", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Write the content of the webpage to a file
			err = os.WriteFile(filename, item.Body, 0644)
			// Verify that the file has been written
			if err == nil {
				w.Write([]byte("\nServing webpage from cache memory \n" + string(jsonStr)))
				return
			}
		} else {
			fmt.Println("Deleted " + req.URL + " from cache memory")
			delete(cache, req.URL)
		}
	}

	// Increment the wait group counter
	wg.Add(1)

	// Acquire a semaphore
	semaphore <- struct{}{}

	// Create a channel to communicate the result of the worker goroutine back to the main goroutine
	resultChan := make(chan error)

	// Start a worker to fetch the webpage
	go func() {
		// Set the retry limit to the minimum of 10 or the retry limit in the request
		retryLimit := 10
		if req.RetryLimit > 0 && req.RetryLimit < retryLimit {
			retryLimit = req.RetryLimit
		}

		// Fetch the webpage
		var resp *http.Response
		for i := 1; i <= retryLimit; i++ {
			resp, err = http.Get(req.URL)
			if err == nil {
				break
			}
			time.Sleep(time.Second)
		}
		if err != nil {
			resultChan <- err
			return
		}
		defer resp.Body.Close()

		// Read the content of the webpage
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			resultChan <- err
			return
		}

		// Create a file with the unique ID as the filename
		file, err := os.Create(filename)
		if err != nil {
			resultChan <- err
			return
		}
		defer file.Close()

		// Download the content of the webpage to the file
		err = os.WriteFile(filename, body, 0644)
		if err != nil {
			resultChan <- err
			return
		}

		// Write the content of the webpage to the cache
		cache[req.URL] = CacheItem{
			Body:      body,
			Timestamp: time.Now(),
			ID:        id,
		}
		fmt.Println("Wrote " + req.URL + " to cache memory")

		// Send the result to the result channel
		resultChan <- nil

	}()

	// Wait for the worker to finish
	err = <-resultChan
	if err != nil {
		http.Error(w, "Error fetching the requested URL", http.StatusInternalServerError)
		return
	}

	// Release the semaphore
	<-semaphore

	// Decrement the wait group counter
	wg.Done()

	// Write the response object to the response
	w.Write([]byte("\nWebpage Successfully Downloaded \n" + string(jsonStr)))

	// Start a goroutine to periodically remove timestamps from cache which are older than 24 hours
	go func() {
		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
			for url, item := range cache {
				if time.Since(item.Timestamp) > cacheExpiryInterval {
					fmt.Println("Deleted " + url + " from cache memory")
					delete(cache, url)
				}
			}
		}
	}()

}

func main() {
	// Initialize the cache map
	cache = make(map[string]CacheItem)

	port := os.Getenv("PORT")
	if port == "" {
		// The PORT environment variable is not set or is not a valid integer.
		// Use the default value of 7771.
		port = "7771"
	}

	http.HandleFunc("/pagesource", downloadPageSource)

	//Listen on the specified port
	fmt.Println("Listening on port: " + port + "...")
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
