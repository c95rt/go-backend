package api

import (
	"net/http"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/gorilla/schema"
	"github.com/mitchellh/mapstructure"
	"github.com/thedevsaddam/govalidator"
)

func InsertAdminUser(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	var opts models.InsertUserOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.InsertUserRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	var err error
	opts.Password, err = helpers.HashPassword(opts.Password)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed hashing password")
		return
	}

	userID, err := ctx.DB.InsertUser(&opts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed inserting user")
		return
	}

	var roles []models.Role
	for _, roleID := range opts.Roles {
		roles = append(roles, models.Role{
			ID: roleID,
		})
	}

	user := models.User{
		ID:        userID,
		Firstname: opts.Firstname,
		Lastname:  opts.Lastname,
		Email:     opts.Email,
		Password:  opts.Password,
		Active:    true,
		Additional: &models.UserAdditional{
			DNI:   opts.DNI,
			Phone: opts.Phone,
		},
		Roles: roles,
	}

	w.WriteJSON(http.StatusOK, user, nil, "")
}

func GetUsers(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	if !userInfo.IsAdmin {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.GetUsersRules,
	}
	v := govalidator.New(validatorOpts)
	errs := v.Validate()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	var opts models.GetUsersOpts
	decoder := schema.NewDecoder()
	decoder.Decode(&opts, r.URL.Query())

	users, err := ctx.DB.GetUsers(&opts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting users")
		return
	}

	if users == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "users not found")
		return
	}

	w.WriteJSON(http.StatusOK, users, nil, "")
}
