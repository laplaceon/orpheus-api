package api

import "time"

type User struct {
	Id        int       `json:"id"`
	Email     string    `json:"email"`
	password  string    `json:"-"`
	Verified  bool      `json:"verified"`
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
	CreditsPerMonth float64 `json:"credits_per_month"`
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
	Amount    float64   `json:"amount"`
	StripeId  string    `json:"stripe_transaction_id"`
	CreatedAt time.Time `json:"created_at"`
}

type HistoryItem struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id"`
	ActionId   int       `json:"action_id"`
	ActionName string    `json:"action_name"`
	Cost       float64   `json:"cost"`
	Status     int       `json:"status"`
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
	Cost   float64 `json:"cost"`
	Length float64 `json:"length"`
}

type ActionCost struct {
	Id       int     `json:"id"`
	ActionId int     `json:"action_id"`
	Name     string  `json:"name"`
	Cost     float64 `json:"cost"`
	Length   float64 `json:"length"`
}
