package main

import "shared/utils"

type PaymentConfig struct {
	DeclineRate           float64
	InsufficientFundsRate float64
	BankTimeoutRate       float64
	BankTimeoutMs         int

	BankHangEnabled bool
	BankHangRate    float64
}

func loadPaymentConfig() PaymentConfig {
	return PaymentConfig{
		DeclineRate:           utils.EnvFloat("PAYMENT_DECLINE_RATE", 0.1),
		InsufficientFundsRate: utils.EnvFloat("PAYMENT_INSUFFICIENT_FUNDS_RATE", 0.05),
		BankTimeoutRate:       utils.EnvFloat("PAYMENT_BANK_TIMEOUT_RATE", 0.05),
		BankTimeoutMs:         utils.EnvInt("PAYMENT_BANK_TIMEOUT_MS", 8000),
		BankHangEnabled:       utils.EnvBool("PAYMENT_BANK_HANG_ENABLED", false),
		BankHangRate:          utils.EnvFloat("PAYMENT_BANK_HANG_RATE", 0.01),
	}
}
