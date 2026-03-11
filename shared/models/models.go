package models

import "time"

type OrderStatus string

const (
	OrderStatusPending    OrderStatus = "pending"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusFailed     OrderStatus = "failed"
)

type Order struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Amount    float64     `json:"amount"`
	Items     []OrderItem `json:"items"`
	Status    OrderStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Error     string      `json:"error,omitempty"`
}

type OrderItem struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateOrderRequest struct {
	UserID string      `json:"user_id"`
	Items  []OrderItem `json:"items"`
}

type PaymentRequest struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
}

type PaymentResponse struct {
	OrderID       string    `json:"order_id"`
	TransactionID string    `json:"transaction_id"`
	Status        string    `json:"status"`
	ProcessedAt   time.Time `json:"processed_at"`
	Error         string    `json:"error,omitempty"`
}
