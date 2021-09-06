package api

import (
	"net/http"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/google/uuid"
	"github.com/thedevsaddam/govalidator"
)

func Login(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	var opts models.LoginOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.LoginRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	user, err := ctx.DB.GetUserLoginByEmail(opts.Email)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting user")
		return
	}

	if user == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "user not found")
		return
	}

	if !helpers.AuthenticateHashedPassword(user.Password, opts.Password) {
		w.WriteJSON(http.StatusNotFound, nil, nil, "user not found")
		return
	}

	user.Token, err = helpers.GenerateToken(user, ctx.Config.JWTSecret)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed generating token")
		return
	}

	w.WriteJSON(http.StatusOK, user, nil, "")
	return
}

func UpdateUserPassword(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	var opts models.UpdateUserPasswordOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.UpdateUserPasswordRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	user, err := ctx.DB.GetUserByRememberToken(opts.Token)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting user")
		return
	}

	if user == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "user not found")
		return
	}

	password, err := helpers.HashPassword(opts.Password)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed hashing password")
		return
	}

	user.Password = password

	err = ctx.DB.UpdateUserPassword(user)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed updating user password")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func SendRememberToken(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	var opts models.SendRememberTokenOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.SendRememberTokenRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	user, err := ctx.DB.GetUserLoginByEmail(opts.Email)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting user")
		return
	}

	if user == nil {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "user not found")
		return
	}

	token, err := uuid.NewUUID()
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed generating remember token")
		return
	}
	err = ctx.DB.UpdateUserRememberToken(user.ID, token.String())
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed updating remember token")
		return
	}

	// SEND EMAIL WITH REMEMBER TOKEN TO REACTIVATE ACCOUNT
	w.WriteJSON(http.StatusNoContent, nil, nil, "")
	return
}
