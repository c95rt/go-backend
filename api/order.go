package api

import (
	"net/http"
	"time"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/db"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
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
