package api

import "time"

type User struct {
	Id        int       `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type ApiKey struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Name      string    `json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `json:"created_at"`
}

type Plan struct {
	Id              int     `json:"id"`
	Name            string  `json:"name"`
	CreditsPerMonth float32 `json:"credits_per_month"`
}

type PlanPurchase struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	PlanId    int       `json:"plan_id"`
	StripeId  string    `json:"stripe_transaction_id"`
	CreatedAt time.Time `json:"created_at"`
}

type CreditPurchase struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	Amount    float32   `json:"amount"`
	StripeId  string    `json:"stripe_transaction_id"`
	CreatedAt time.Time `json:"created_at"`
}

type HistoryItem struct {
	Id        int       `json:"id"`
	UserId    int       `json:"user_id"`
	ActionId  int       `json:"action_id"`
	InputSize float32   `json:"input_size"`
	CreatedAt time.Time `json:"created_at"`
}

type Action struct {
	Id   int     `json:"id"`
	Name string  `json:"name"`
	Cost float32 `json:"cost"`
}
