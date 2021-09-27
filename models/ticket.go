package models

import "github.com/thedevsaddam/govalidator"

type GetTicketsOpts struct {
	EventFrom  string `schema:"event_from"`
	EventTo    string `schema:"event_to"`
	LimitFrom  int    `schema:"limit_from"`
	LimitTo    int    `schema:"limit_to"`
	TicketUUID string `schema:"ticket_uuid"`
}

var GetTicketsRules = govalidator.MapData{
	"event_from":  []string{"date_ISO8601"},
	"event_to":    []string{"date_ISO8601"},
	"limit_from":  []string{"numeric"},
	"limit_to":    []string{"numeric"},
	"ticket_uuid": []string{"alpha"},
}

type GetTicketsStruct struct {
	Tickets []Ticket `json:"tickets,omitempty"`
	Total   int      `json:"total"`
}
