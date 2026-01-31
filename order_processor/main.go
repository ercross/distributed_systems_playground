// order-processor/main.go
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
	http.HandleFunc("/process", handleProcess)

	log.Println("Order Processor running on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Processing order: %+v", order)

	// Call payment service
	body, _ := json.Marshal(order)
	resp, err := http.Post("http://payment-service:8082/payment", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("ERROR: Payment service unreachable: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("ERROR: Payment failed with status %d", resp.StatusCode)
		http.Error(w, "payment failed", resp.StatusCode)
		return
	}

	log.Println("Order processed and paid successfully")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Order processed\n")
}
