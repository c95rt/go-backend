package models

import "time"

type Payment struct {
	ID           int            `json:"id,omitempty"`
	Method       *PaymentMethod `json:"method,omitempty"`
	Amount       int            `json:"amount,omitempty"`
	User         *User          `json:"user,omitempty"`
	PreferenceID string         `json:"preference_id,omitempty"`
	Order        *Order         `json:"order,omitempty"`
	Status       *PaymentStatus `json:"payment_status,omitempty"`
	Created      time.Time      `json:"created"`
	Updated      time.Time      `json:"updated"`
}

type PaymentMethod struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type PaymentStatus struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
