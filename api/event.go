package api

import (
	"fmt"
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

func InsertEvents(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsCashier && !userInfo.IsAdmin {
		w.WriteJSON(http.StatusForbidden, nil, nil, "must be admin")
		return
	}

	timeLocation, err := time.LoadLocation("America/Santiago")
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed loading time location")
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
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	for _, date := range opts.Dates {
		if _, err := time.Parse(db.ConstLayoutDate, date.Date); err != nil {
			w.WriteJSON(http.StatusBadRequest, nil, err, fmt.Sprintf("date must be like %s", db.ConstLayoutDate))
			return
		}

		if len(date.Times) == 0 {
			w.WriteJSON(http.StatusBadRequest, nil, nil, "the times field is required")
			return
		}

		for _, eventTime := range date.Times {
			startDateTime, err := time.ParseInLocation(db.ConstLayoutDateTime, fmt.Sprintf("%s %s", date.Date, eventTime.StartTime), timeLocation)
			if err != nil {
				w.WriteJSON(http.StatusBadRequest, nil, err, fmt.Sprintf("end_time must be like %s", db.ConstLayoutTime))
				return
			}
			endDateTime, err := time.ParseInLocation(db.ConstLayoutDateTime, fmt.Sprintf("%s %s", date.Date, eventTime.EndTime), timeLocation)
			if err != nil {
				w.WriteJSON(http.StatusBadRequest, nil, err, fmt.Sprintf("end_time must be like %s", db.ConstLayoutTime))
				return
			}
			if endDateTime.Before(startDateTime) {
				w.WriteJSON(http.StatusBadRequest, nil, nil, fmt.Sprintf("end of event must be after start"))
				return
			}
			if endDateTime.Before(time.Now()) {
				w.WriteJSON(http.StatusBadRequest, nil, nil, fmt.Sprintf("end of event must be after now"))
				return
			}
		}
	}

	if err := ctx.DB.InsertEvents(&opts); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed inserting events")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func GetEvents(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetEventsRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	var opts models.GetEventsOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	events, err := ctx.DB.GetEvents(&opts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting events")
		return
	}

	if events == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "events not found")
		return
	}

	w.WriteJSON(http.StatusOK, events, nil, "")
}
