package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/db"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/thedevsaddam/govalidator"
)

func InsertOrder(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	// timeLocation, err := time.LoadLocation("America/Santiago")
	// if err != nil {
	// 	w.WriteJSON(http.StatusInternalServerError, nil, err, "failed loading time location")
	// }

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	var opts models.InsertOrderOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.InsertOrderRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	user, err := ctx.DB.GetUserByID(opts.UserID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting user")
		return
	}

	if user == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "user not found")
		return
	}

	if !user.HasRole(db.ConstRoles.Admin) {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "order owner must be client")
		return
	}

	ticketsByEventID := make(map[int]int)
	var finalEventIDs []int
	for _, eventID := range opts.EventIDs {
		if _, ok := ticketsByEventID[eventID]; !ok {
			finalEventIDs = append(finalEventIDs, eventID)
		}
		ticketsByEventID[eventID] += 1
	}

	events, err := ctx.DB.GetEventsByIDs(finalEventIDs)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting events")
		return
	}

	if len(events) != len(finalEventIDs) {
		w.WriteJSON(http.StatusNotFound, nil, nil, "not all events were found")
		return
	}

	for _, event := range events {
		if event.EndDateTime.Before(time.Now()) {
			w.WriteJSON(http.StatusBadRequest, nil, nil, "event is already finished")
			return
		}
	}

	order, err := ctx.DB.InsertOrder(userInfo.ID, opts.UserID, events, ticketsByEventID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed inserting order")
		return
	}

	// Generate PDFs and send emails

	w.WriteJSON(http.StatusOK, order, nil, "")
}

func GetOrders(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier && !userInfo.IsClient {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetOrdersRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	var opts models.GetOrdersOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		opts.ClientIDs = []int{userInfo.ID}
		opts.UserIDs = []int{}
	}
}

func DeleteOrderTicket(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	orderTicket := vars["order_ticket"]
	orderTicketArray := strings.Split(orderTicket, "x")
	if len(orderTicketArray) != 2 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "invalid order ticket")
		return
	}
	orderIDString := orderTicketArray[0]
	ticketIDString := orderTicketArray[1]
	if len(orderIDString) == 0 || len(ticketIDString) == 0 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "invalid order ticket")
		return
	}

	orderID, err := strconv.Atoi(orderIDString)
	if err != nil {
		w.WriteJSON(http.StatusBadRequest, nil, err, "invalid order")
		return
	}
	ticketID, err := strconv.Atoi(ticketIDString)
	if err != nil {
		w.WriteJSON(http.StatusBadRequest, nil, err, "invalid ticket")
		return
	}

	order, err := ctx.DB.GetOrderByOrderIDAndTicketID(orderID, ticketID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting order")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "order not found")
		return
	}

	if len(order.Tickets) == 0 {
		w.WriteJSON(http.StatusNotFound, nil, nil, "ticket not found")
		return
	}

	if !order.Paid {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "order not paid")
		return
	}

	event := order.Tickets[0].Event
	if event == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "event not found")
		return
	}

	if !time.Now().After(event.StartDateTime) && !time.Now().Equal(event.StartDateTime) {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "event not started")
		return
	}

	if !time.Now().Before(event.EndDateTime) {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "event finished")
		return
	}

	if err := ctx.DB.DeleteTicket(ticketID); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed deleting ticket")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}
