package api

import (
	"net/http"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/server"
)

// HealthcheckHandler indicates the service's healthy
func HealthcheckHandler(_ *config.AppContext, w *middlewares.ResponseWriter, _ *http.Request) {
	w.String(http.StatusOK, "OK")
}

// GetRoutes ...
func GetRoutes() []*server.Route {
	return []*server.Route{
		{Path: "/healthcheck", Methods: []string{"GET", "HEAD"}, Handler: HealthcheckHandler, IsProtected: false},

		// Auth
		{Path: "/auth/login", Methods: []string{"POST", "HEAD"}, Handler: Login, IsProtected: false},
		{Path: "/auth/password", Methods: []string{"PUT", "HEAD"}, Handler: UpdateUserPassword, IsProtected: false},
		{Path: "/auth/token", Methods: []string{"POST", "HEAD"}, Handler: SendRememberToken, IsProtected: false},

		// User
		{Path: "/user/admin", Methods: []string{"POST", "HEAD"}, Handler: InsertAdminUser, IsProtected: true},
		{Path: "/user", Methods: []string{"POST", "HEAD"}, Handler: InsertUser, IsProtected: false},
		{Path: "/user/{id:[0-9]+}", Methods: []string{"PUT", "HEAD"}, Handler: UpdateUser, IsProtected: true},
		{Path: "/user/{id:[0-9]+}", Methods: []string{"GET", "HEAD"}, Handler: GetUser, IsProtected: true},
		{Path: "/user", Methods: []string{"GET", "HEAD"}, Handler: GetUsers, IsProtected: true},

		// Event
		{Path: "/event", Methods: []string{"POST", "HEAD"}, Handler: InsertEvents, IsProtected: true},
		{Path: "/event", Methods: []string{"GET", "HEAD"}, Handler: GetEvents, IsProtected: false},
		{Path: "/event/{id:[0-9]+}", Methods: []string{"GET", "HEAD"}, Handler: GetEvent, IsProtected: false},

		// Order
		{Path: "/order", Methods: []string{"POST", "HEAD"}, Handler: InsertOrder, IsProtected: true},
		{Path: "/order", Methods: []string{"GET", "HEAD"}, Handler: GetOrders, IsProtected: true},
		{Path: "/order/{id:[0-9]+}/pdf", Methods: []string{"GET", "HEAD"}, Handler: GetOrderPDF, IsProtected: true},
		{Path: "/order/{id:[0-9]+}", Methods: []string{"PATCH", "HEAD"}, Handler: UseOrder, IsProtected: true},
		{Path: "/order/{id:[0-9]+}", Methods: []string{"PUT", "HEAD"}, Handler: UpdateOrder, IsProtected: true},
		{Path: "/sales", Methods: []string{"GET", "HEAD"}, Handler: GetSalesSummary, IsProtected: true},
		{Path: "/sales/cashier", Methods: []string{"GET", "HEAD"}, Handler: GetCashierSummary, IsProtected: true},

		// Payment
		{Path: "/payment/{order_id:[0-9]+}/mercadopago", Methods: []string{"POST", "HEAD"}, Handler: InsertPaymentMercadoPago, IsProtected: true},
		{Path: "/payment/{order_id:[0-9]+}/cashier", Methods: []string{"POST", "HEAD"}, Handler: InsertPaymentCashier, IsProtected: true},
		{Path: "/payment/mercadopago", Methods: []string{"POST", "HEAD"}, Handler: UpdatePaymentMercadoPago, IsProtected: false},
	}
}
