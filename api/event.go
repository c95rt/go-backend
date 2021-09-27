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

func InsertEvents(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsCashier && !userInfo.IsAdmin {
		w.Write(http.StatusForbidden, nil, nil, middlewares.Responses.InvalidRoles)
		return
	}

	timeLocation, err := time.LoadLocation("America/Santiago")
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	var opts models.InsertEventsOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.InsertEventsRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.Write(http.StatusBadRequest, errs, nil, middlewares.Responses.FailedValidations)
		return
	}

	for _, date := range opts.Dates {
		if _, err := time.Parse(db.ConstLayoutDate, date.Date); err != nil {
			w.Write(http.StatusBadRequest, nil, err, middlewares.Responses.FailedValidations)
			return
		}

		if len(date.Times) == 0 {
			w.Write(http.StatusBadRequest, nil, nil, middlewares.Responses.TimeFieldRequired)
			return
		}

		for _, eventTime := range date.Times {
			startDateTime, err := time.ParseInLocation(db.ConstLayoutDateTime, fmt.Sprintf("%s %s", date.Date, eventTime.StartTime), timeLocation)
			if err != nil {
				w.Write(http.StatusBadRequest, nil, err, middlewares.Responses.FailedValidations)
				return
			}
			endDateTime, err := time.ParseInLocation(db.ConstLayoutDateTime, fmt.Sprintf("%s %s", date.Date, eventTime.EndTime), timeLocation)
			if err != nil {
				w.Write(http.StatusBadRequest, nil, err, middlewares.Responses.FailedValidations)
				return
			}
			if endDateTime.Before(startDateTime) {
				w.Write(http.StatusBadRequest, nil, nil, middlewares.Responses.EndTimeBeforeStartTime)
				return
			}
			if endDateTime.Before(time.Now()) {
				w.Write(http.StatusBadRequest, nil, nil, middlewares.Responses.EndTimeBeforeNow)
				return
			}
		}
	}

	if err := ctx.DB.InsertEvents(&opts); err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func GetEvents(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetEventsRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.Write(http.StatusBadRequest, errs, nil, middlewares.Responses.FailedValidations)
		return
	}

	var opts models.GetEventsOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	events, err := ctx.DB.GetEvents(&opts)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	w.WriteJSON(http.StatusOK, events, nil, "")
}

func GetEvent(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

	vars := mux.Vars(r)
	eventID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	event, err := ctx.DB.GetEventByID(eventID)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	if event == nil {
		w.Write(http.StatusNotFound, nil, nil, middlewares.Responses.EventNotFound)
		return
	}

	w.WriteJSON(http.StatusOK, event, nil, "")
}
