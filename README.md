# Webpage Downloader

This is a server endpoint that takes the URL of a webpage as input, fetches the webpage, and downloads it as a file on the local file system.

## Features

- The server accepts a retry limit as a parameter. It retries a maximum of 10 times or the retry limit, whichever is lower, before either successfully downloading the webpage or marking the page as a failure.
- If the webpage has already been requested in the last 24 hours, it is served from the local cache.
- The server has a pool of workers that do the work of downloading the requested webpage. This allows the server to handle a large number of concurrent requests while still limiting the number of actual requests to download the webpages.

## Usage

To start the server, run the following command:

```go
go run main.go
```


To send a request to the server, use a tool such as `curl` to send a POST request to the `/pagesource` endpoint with a JSON body containing the URL of the webpage you want to download.

Here's an example of how you can send a request using `curl`:

```bash
curl --location --request POST 'http://localhost:5000/pagesource' \
--header 'Content-Type: application/json' \
--data-raw '{
   "url": "https://google.com",
   "retryLimit": 3
}'
```


The server will respond with a JSON object containing the unique ID of the request, the URL of the requested webpage, and the source URL of the downloaded webpage.

Here's an example of the server's response:

```bash
{
"id": "c33357d82b6105c59e3089f2b70af7f8",
"url": "https://google.com",
"sourceUrl": "/files/c33357d82b6105c59e3089f2b70af7f8.html"
}
```

The downloaded webpage will be saved to the local file system as a file with the name `[ID].html`, where `[ID]` is the unique ID of the request. The file will be saved in the `files` directory.

## Configuration

The server can be configured using the following environment variables:

- `PORT`: the port number on which the server will listen for requests (default: 5000)
- `RETRY_LIMIT`: the maximum number of retries for each request (default: 10)
- `WORKER_POOL_SIZE`: the size of the worker pool (default: 10)

## Dependencies

The server has the following dependencies:

- [Go](https://golang.org/)

