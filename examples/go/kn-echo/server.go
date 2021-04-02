package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST requests are supported.", http.StatusNotFound)
		return
	}

	w.WriteHeader(204)

	cloudEvent := extractCloudEvent(r.Body)
	bytes, err := json.Marshal(cloudEvent)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	log.Printf("***cloud event*** %s", bytes)
}

func extractCloudEvent(reqBody io.ReadCloser) map[string]interface{} {
	origEvent := make(map[string]interface{})
	json.NewDecoder(reqBody).Decode(&origEvent)

	// Reorganize origEvent to have two keys: attributes and data, and copy to
	// cloudEvent map.
	cloudEvent := make(map[string]interface{})

	attr := make(map[string]interface{})
	// Pick out the attributes.
	for k, v := range origEvent {
		if k != "data" {
			attr[k] = v
		}
	}
	cloudEvent["attributes"] = attr
	cloudEvent["data"] = origEvent["data"]

	return cloudEvent
}

func main() {
	log.Println("Starting server at port 8080.")
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
