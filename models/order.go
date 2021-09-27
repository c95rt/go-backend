package models

import (
	"time"

	"github.com/thedevsaddam/govalidator"
)

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

type UpdateTicketOpts struct {
	EventID int `json:"event_id"`
}

var UpdateTicketRules = govalidator.MapData{
	"event_id": []string{"required", "numeric"},
}

type Order struct {
	ID      int       `json:"id,omitempty"`
	User    *User     `json:"user,omitempty"`
	Client  *User     `json:"client,omitempty"`
	Tickets []Ticket  `json:"tickets,omitempty"`
	Price   int       `json:"price"`
	Payment *Payment  `json:"payment,omitempty"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

type Ticket struct {
	ID     int    `json:"id,omitempty"`
	UUID   string `json:"uuid,omitempty"`
	Event  *Event `json:"event,omitempty"`
	Used   int    `json:"used"`
	Paid   *bool  `json:"paid,omitempty"`
	Client *User  `json:"client,omitempty"`
}

type TicketHTML struct {
	ID                 int
	Firstname          string
	Lastname           string
	EventStartDateTime string
	EventEndDateTime   string
	Price              int
	Image              string
}

type TicketPDF struct {
	URL string `json:"url"`
}

type OrderHTML struct {
	ID            int
	Firstname     string
	Lastname      string
	PaymentMethod string
	OrderPrice    int
}
