package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Request struct {
	URL        string `json:"url"`
	RetryLimit int    `json:"retryLimit"`
}

type CacheItem struct {
	Body      []byte
	Timestamp time.Time
}

var cache map[string]CacheItem

const cacheExpiryInterval = 24 * time.Hour

func downloadPageSource(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request.", http.StatusBadRequest)
		return
	}

	// Check if the webpage is in the cache
	if item, ok := cache[req.URL]; ok {
		// Check if the webpage was requested within the last 24 hours
		if time.Since(item.Timestamp) < cacheExpiryInterval {
			// Write the content of the webpage to a file
			err = os.WriteFile("webpage.html", item.Body, 0644)
			// Verify that the file has been written
			if err == nil {
				w.Write([]byte("Serving webpage from cache memory"))
				return
			}
		} else {
			fmt.Println("Deleted " + req.URL + " from cache memory")
			delete(cache, req.URL)
		}
	}

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
		http.Error(w, "Error fetching the requested URL after "+strconv.Itoa(retryLimit)+" attempts", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the content of the webpage
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading the requested webpage.", http.StatusInternalServerError)
		return
	}

	// Download the content of the webpage to a file
	err = os.WriteFile("webpage.html", body, 0644)
	if err != nil {
		http.Error(w, "Error downloading the requested webpage.", http.StatusInternalServerError)
		return
	}

	// Write the content of the webpage to the cache
	cache[req.URL] = CacheItem{
		Body:      body,
		Timestamp: time.Now(),
	}
	fmt.Println("Wrote " + req.URL + " to cache memory")

	w.Write([]byte("Webpage Successfully Downloaded"))

	// Start a goroutine to periodically check the timestamps of the cache items and remove those older than 24 hours
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

	http.HandleFunc("/pagesource", downloadPageSource)
	fmt.Println("Listening on :5000...")
	http.ListenAndServe(":5000", nil)
}
