package main

import (
	"encoding/json"
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

func downloadPageSource(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request.", http.StatusBadRequest)
		return
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

	w.Write([]byte("Webpage Successfully Downloaded"))
}

func main() {
	http.HandleFunc("/pagesource", downloadPageSource)
	http.ListenAndServe(":5000", nil)
}
