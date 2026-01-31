package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type Order struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/payment", handlePayment)

	log.Println("Payment Service running on :8082 (50% failure rate)")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func handlePayment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Payment request for user=%s amount=%.2f", order.UserID, order.Amount)

	// Randomly fail 50% of the time
	if rand.Float64() < 0.5 {
		log.Printf("PAYMENT FAILED for user=%s", order.UserID)
		http.Error(w, "payment processing failed", http.StatusInternalServerError)
		return
	}

	log.Printf("PAYMENT SUCCESS for user=%s amount=%.2f", order.UserID, order.Amount)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Payment processed\n")
}
