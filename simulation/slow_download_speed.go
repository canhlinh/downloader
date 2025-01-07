package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func slowDownloadHandler(w http.ResponseWriter, r *http.Request) {
	// Path to the file you want to serve
	filePath := "SampleVideo_1280x720_20mb.mp4"
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found.", http.StatusNotFound)
		return
	}
	defer file.Close()

	// Get file info for size (useful for HEAD requests)
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get file info.", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodHead {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileInfo.Name()))
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
		return
	}

	// Set headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filePath))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Buffer size for reading chunks
	buffer := make([]byte, 1024)
	for {
		// Read a chunk of the file
		n, err := file.Read(buffer)
		if n > 0 {
			// Write the chunk to the response
			_, writeErr := w.Write(buffer[:n])
			if writeErr != nil {
				fmt.Println("Error writing to response:", writeErr)
				return
			}
			// Flush the response to send the chunk
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			// Add delay to simulate slow download
			time.Sleep(500 * time.Millisecond)
		}
		if err != nil {
			break
		}
	}
}

func main() {
	http.HandleFunc("/SampleVideo_1280x720_20mb.mp4", slowDownloadHandler)

	fmt.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
