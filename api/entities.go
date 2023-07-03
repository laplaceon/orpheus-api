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
	ExpiryDays      int     `json:"expiry_days"`
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
	Id         int       `json:"id"`
	UserId     int       `json:"user_id"`
	ActionId   int       `json:"action_id"`
	ActionName string    `json:"action_name"`
	Cost       float32   `json:"cost"`
	CreatedAt  time.Time `json:"created_at"`
}

type GeneratedItem struct {
	Id        int       `json:"id"`
	HistoryId int       `json:"history_id"`
	PlanId    int       `json:"plan_id"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

type Action struct {
	Id     int     `json:"id"`
	Name   string  `json:"name"`
	Cost   float32 `json:"cost"`
	Length float32 `json:"length"`
}

// type ActionCost struct {
// 	Id        int       `json:"id"`
// 	ActionId  int       `json:"action_id"`
// 	Cost      float32   `json:"cost"`
// 	Length    float32   `json:"length"`
// 	CreatedAt time.Time `json:"created_at"`
// }
