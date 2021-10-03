package models

import (
	"time"

	"github.com/thedevsaddam/govalidator"
)

type InsertEventsOpts struct {
	TypeID int                   `json:"type_id"`
	Dates  []InsertEventDateOpts `json:"dates"`
}

var InsertEventsRules = govalidator.MapData{
	"dates":   []string{"required"},
	"type_id": []string{"required", "numeric"},
}

type InsertEventDateOpts struct {
	Date  string                    `json:"date"`
	Times []InsertEventDateTimeOpts `json:"times"`
}

type InsertEventDateTimeOpts struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Price     int    `json:"price"`
}

type GetEventsOpts struct {
	Date      string `schema:"date"`
	TypeID    int    `schema:"type_id"`
	LimitFrom int    `schema:"limit_from"`
	LimitTo   int    `schema:"limit_to"`
}

var GetEventsRules = govalidator.MapData{
	"date":       []string{"date_ISO8601"},
	"type_id":    []string{"numeric"},
	"limit_from": []string{"numeric"},
	"limit_to":   []string{"numeric"},
}

type Event struct {
	ID            int        `json:"id,omitempty"`
	Type          *EventType `json:"type,omitempty"`
	StartDateTime time.Time  `json:"start_date_time"`
	EndDateTime   time.Time  `json:"end_date_time"`
	Price         int        `json:"price"`
	Created       time.Time  `json:"created"`
	Updated       time.Time  `json:"updated"`
}

type EventType struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type EventsStruct struct {
	Events []Event `json:"events"`
	Total  int     `json:"total"`
}
