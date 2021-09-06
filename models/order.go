package models

import "github.com/thedevsaddam/govalidator"

type InsertOrderOpts struct {
	UserID   int   `json:"user_id"`
	EventIDs []int `json:"event_ids"`
}

var InsertOrderRules = govalidator.MapData{
	"user_id":   []string{"required", "numeric"},
	"event_ids": []string{"required", "array_int"},
}

type Order struct {
	ID      int      `json:"id,omitempty"`
	User    *User    `json:"user,omitempty"`
	Client  *User    `json:"client,omitempty"`
	Tickets []Ticket `json:"tickets,omitempty"`
	Price   int      `json:"price"`
}

type Ticket struct {
	ID    int    `json:"id,omitempty"`
	Event *Event `json:"event,omitempty"`
}
