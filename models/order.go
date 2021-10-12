package models

import (
	"time"

	"github.com/thedevsaddam/govalidator"
)

type InsertOrdersOpts struct {
	UserID  int `json:"user_id"`
	EventID int `json:"event_id"`
	Tickets int `json:"tickets"`
}

var InsertOrdersRules = govalidator.MapData{
	"user_id":  []string{"required", "numeric"},
	"event_id": []string{"required", "numeric"},
	"tickets":  []string{"required", "numeric"},
}

type UpdateOrderOpts struct {
	EventID int `json:"event_id"`
}

var UpdateOrderRules = govalidator.MapData{
	"event_id": []string{"required", "numeric"},
}

type GetOrdersOpts struct {
	EventFrom     string `schema:"event_from"`
	EventTo       string `schema:"event_to"`
	LimitFrom     int    `schema:"limit_from"`
	LimitTo       int    `schema:"limit_to"`
	TransactionID string `schema:"transaction_id"`
	EventTypeID   int    `schema:"event_type_id"`
	Paid          *bool  `schema:"paid"`
	ClientID      int    `schema:"client_id"`
}

var GetOrdersRules = govalidator.MapData{
	"event_from":     []string{"date_ISO8601"},
	"event_to":       []string{"date_ISO8601"},
	"limit_from":     []string{"numeric"},
	"limit_to":       []string{"numeric"},
	"transaction_id": []string{"alpha"},
	"event_type_id":  []string{"numeric"},
	"paid":           []string{"bool"},
	"client_id":      []string{"numeric"},
}

type GetCashierSummaryOpts struct {
	DateFrom  string `schema:"date_from"`
	DateTo    string `schema:"date_to"`
	CashierID int    `schema:"cashier_id"`
}

var GetCashierSummaryRules = govalidator.MapData{
	"date_from":  []string{"date_ISO8601", "required"},
	"date_to":    []string{"date_ISO8601", "required"},
	"cashier_id": []string{"numeric", "required"},
}

type Order struct {
	ID            int       `json:"id,omitempty"`
	User          *User     `json:"user,omitempty"`
	Client        *User     `json:"client,omitempty"`
	Event         *Event    `json:"event,omitempty"`
	TransactionID string    `json:"transaction_id"`
	Tickets       int       `json:"tickets"`
	Price         int       `json:"price"`
	Payment       *Payment  `json:"payment,omitempty"`
	Paid          *bool     `json:"paid,omitempty"`
	Used          *bool     `json:"used,omitempty"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}

type OrderPDFHTML struct {
	ID            int
	Firstname     string
	Lastname      string
	Date          string
	EventType     string
	Price         int
	Image         string
	TransactionID string
	Tickets       int
}

type OrderPDF struct {
	URL string `json:"url"`
}

type OrderHTML struct {
	ID            int
	Firstname     string
	Lastname      string
	PaymentMethod string
	OrderPrice    int
	TransactionID string
	Tickets       int
	Date          string
}

type GetOrdersStruct struct {
	Orders []Order `json:"orders,omitempty"`
	Total  int     `json:"total"`
}

type SalesSummary struct {
	CurrentDay         int64                       `json:"current_day"`
	CurrentMonth       int64                       `json:"current_month"`
	CurrentYear        int64                       `json:"current_year"`
	MonthlyCurrentYear []MonthlySalesSummaryDetail `json:"monthly_current_year"`
	MonthlyLastYear    []MonthlySalesSummaryDetail `json:"monthly_last_year"`
}

type MonthlySalesSummaryDetail struct {
	Month string `json:"month"`
	Year  int    `json:"year"`
	Total int64  `json:"total"`
}

type DailySales struct {
	Date  time.Time
	Total int64
}

type CashierSummary struct {
	User         *User                       `json:"user,omitempty"`
	TotalSales   int64                       `json:"total_sales"`
	TotalUses    int64                       `json:"total_uses"`
	MonthlySales []MonthlySalesSummaryDetail `json:"monthly_sales"`
	MonthlyUses  []MonthlySalesSummaryDetail `json:"monthly_uses"`
}

type CashierMonthlySales struct {
	User       *User
	TotalSales int64
	TotalUses  int64
	Date       time.Time
}
