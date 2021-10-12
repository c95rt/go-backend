package api

import (
	"net/http"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/gorilla/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/thedevsaddam/govalidator"
)

func InsertCamping(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	// timeLocation, err := time.LoadLocation("America/Santiago")
	// if err != nil {
	// 	w.WriteJSON(http.StatusInternalServerError, nil, err, "failed loading time location")
	// }

	var opts models.InsertCampingOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.InsertCampingRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	event, err := ctx.DB.GetEventByID(opts.EventID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	if event == nil {
		w.WriteJSON(http.StatusNotFound, nil, err, "Evento no encontrado")
		return
	}

	camping, err := ctx.DB.InsertCamping(userInfo.ID, event.ID, opts.Tickets, event.Price)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	camping.Event = event
	camping.Price = event.Price * opts.Tickets

	w.WriteJSON(http.StatusOK, camping, nil, "")
}

func GetCampings(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin && !userInfo.IsCashier && !userInfo.IsClient {
		w.WriteJSON(http.StatusForbidden, nil, nil, "Rol inválido")
		return
	}

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetCampingsRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	var opts models.GetCampingsOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	campings, err := ctx.DB.GetCampings(&opts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "Error del servidor")
		return
	}

	w.WriteJSON(http.StatusOK, campings, nil, "")
}
