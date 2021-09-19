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

type GetOrdersOpts struct {
	CreatedFrom string `schema:"created_from"`
	CreatedTo   string `schema:"created_to"`
	EventFrom   string `schema:"event_from"`
	EventTo     string `schema:"event_to"`
	UserIDs     []int  `schema:"user_ids"`
	ClientIDs   []int  `schema:"client_ids"`
	LimitFrom   int    `schema:"limit_from"`
	LimitTo     int    `schema:"limit_to"`
}

var GetOrdersRules = govalidator.MapData{
	"created_from": []string{"date_ISO8601"},
	"created_to":   []string{"date_ISO8601"},
	"event_from":   []string{"date_ISO8601"},
	"event_to":     []string{"date_ISO8601"},
	"user_ids":     []string{"array_int"},
	"client_ids":   []string{"array_int"},
	"limit_from":   []string{"numeric"},
	"limit_to":     []string{"numeric"},
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
