package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	port := "8001"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	name := fmt.Sprintf("Backend-%s", port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(time.Duration(50+port[len(port)-1]*10) * time.Millisecond)

		fmt.Fprintf(w, "✅ Response from %s\n", name)
		fmt.Printf("[%s] Request: %s %s\n", name, r.Method, r.URL.Path)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"server": "%s", "timestamp": "%s"}`, name, time.Now().Format(time.RFC3339))
	})

	fmt.Printf("🚀 %s running on http://localhost:%s\n", name, port)
	http.ListenAndServe(":"+port, nil)
}
