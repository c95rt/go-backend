package api

import (
	"fmt"
	"net/http"
	"strconv"
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

func GetTicketByUUID(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	ticketUUID := vars["uuid"]

	order, err := ctx.DB.GetOrderByTicketUUID(ticketUUID)
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

	w.WriteJSON(http.StatusOK, order, nil, "")
}

func GetTicket(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	ticketID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed parsing ticket ID")
		return
	}

	fmt.Println("aqui")
	order, err := ctx.DB.GetOrderByTicketID(ticketID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting order")
		return
	}
	fmt.Println("aqui2")

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "order not found")
		return
	}

	if len(order.Tickets) == 0 {
		w.WriteJSON(http.StatusNotFound, nil, nil, "ticket not found")
		return
	}

	w.WriteJSON(http.StatusOK, order, nil, "")
}

func UseTicket(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	ticketUUID := vars["uuid"]

	order, err := ctx.DB.GetOrderByTicketUUID(ticketUUID)
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

	if order.Tickets[0].Used == 1 {
		w.WriteJSON(http.StatusNotFound, nil, nil, "ticket used")
		return
	}

	if order.Payment == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "order not paid")
		return
	}

	fmt.Printf("%+v", order.Payment)

	if order.Payment.Status == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "order not paid")
		return
	}

	if order.Payment.Status.ID != db.ConstPaymentStatuses.Approved.ID {
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

	if err := ctx.DB.UseTicket(order.Tickets[0].ID); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed deleting ticket")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func UpdateTicket(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	vars := mux.Vars(r)
	ticketID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed parsing ticket ID")
		return
	}

	var opts models.UpdateTicketOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.UpdateTicketRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	order, err := ctx.DB.GetOrderByTicketID(ticketID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting order")
		return
	}

	if order == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "order not found")
		return
	}

	if len(order.Tickets) == 0 {
		w.WriteJSON(http.StatusNotFound, nil, err, "ticket not found")
		return
	}

	if order.Tickets[0].Used == 1 {
		w.WriteJSON(http.StatusBadRequest, nil, err, "ticket already used")
		return
	}

	if err := ctx.DB.UpdateTicket(ticketID, opts.EventID); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed updating ticket")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func GetTickets(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier {
		w.Write(http.StatusForbidden, nil, nil, middlewares.Responses.InvalidRoles)
		return
	}

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetTicketsRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.Write(http.StatusBadRequest, errs, nil, middlewares.Responses.FailedValidations)
		return
	}

	var opts models.GetTicketsOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	tickets, err := ctx.DB.GetTickets(&opts)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	w.WriteJSON(http.StatusOK, tickets, nil, "")
}
