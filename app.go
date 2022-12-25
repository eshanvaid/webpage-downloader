package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
)

type Request struct {
	URL string `json:"url"`
}

func downloadPageSource(w http.ResponseWriter, r *http.Request) {
	// Parsing the request body
	var req Request
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request.", http.StatusBadRequest)
		return
	}

	// Fetching the webpage
	resp, err := http.Get(req.URL)
	if err != nil {
		http.Error(w, "Error fetching the requested URL.", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Reading the content of the webpage
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading the requested webpage.", http.StatusInternalServerError)
		return
	}

	// Downloading the content of the webpage to a file
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
