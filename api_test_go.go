package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendSwiftCode(t *testing.T) {
	// Define test input data
	data := map[string]interface{}{
		"address":       "ul. Przykładowa 123",
		"bankname":      "Bank Polska",
		"countryiso2":   "PL",
		"countryname":   "Polska",
		"isheadquarter": true,
		"swiftcode":     "POLUPLXXX",
	}

	// Marshal the data into JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("Failed to marshal data: %v", err)
	}

	// Mock the HTTP POST method
	http.HandleFunc("/v1/swift-codes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected method POST, got %s", r.Method)
		}

		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Failed to decode response body: %v", err)
		}

		// Check if the correct data was sent
		if body["swiftcode"] != "POLUPLXXX" {
			t.Errorf("Expected swiftcode to be 'POLUPLXXX', got %v", body["swiftcode"])
		}

		// Respond with a success message
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "SWIFT code successfully added."})
	})

	// Create a test server
	ts := httptest.NewServer(nil)
	defer ts.Close()

	// Send the POST request to our test server
	resp, err := http.Post(ts.URL+"/v1/swift-codes", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", resp.Status)
	}
}
