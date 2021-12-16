package models

import (
	"time"

	"github.com/thedevsaddam/govalidator"
)

type InsertCampingOpts struct {
	EventID int `json:"event_id"`
	Tickets int `json:"tickets"`
}

var InsertCampingRules = govalidator.MapData{
	"event_id": []string{"required", "numeric"},
	"tickets":  []string{"required", "numeric"},
}

type GetCampingsOpts struct {
	EventFrom     string `schema:"event_from"`
	EventTo       string `schema:"event_to"`
	LimitFrom     int    `schema:"limit_from"`
	LimitTo       int    `schema:"limit_to"`
	TransactionID string `schema:"transaction_id"`
	ClientID      int    `schema:"client_id"`
}

var GetCampingsRules = govalidator.MapData{
	"event_from":     []string{"date_ISO8601"},
	"event_to":       []string{"date_ISO8601"},
	"limit_from":     []string{"numeric"},
	"limit_to":       []string{"numeric"},
	"client_id":      []string{"numeric"},
	"transaction_id": []string{},
}

type Camping struct {
	ID            int       `json:"id,omitempty"`
	Client        *User     `json:"client,omitempty"`
	Event         *Event    `json:"event,omitempty"`
	TransactionID string    `json:"transaction_id"`
	Tickets       int       `json:"tickets"`
	Price         int       `json:"price"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

type GetCampingsStruct struct {
	Campings []Camping `json:"campings,omitempty"`
	Total    int       `json:"total"`
}
