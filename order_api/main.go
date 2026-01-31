// order-api/main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Order struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

func main() {
	http.HandleFunc("/orders", handleOrder)

	log.Println("Order API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received order: %+v", order)

	// Forward to processor
	body, _ := json.Marshal(order)
	resp, err := http.Post("http://order-processor:8081/process", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("ERROR: Failed to send to processor: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Processor returned status %d", resp.StatusCode)
		http.Error(w, "processor failed", resp.StatusCode)
		return
	}

	log.Println("Order forwarded successfully")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Order accepted\n")
}
