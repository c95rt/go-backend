package api

import (
	"fmt"
	"net/http"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/google/uuid"
	"github.com/thedevsaddam/govalidator"
)

func Login(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

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
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	if user == nil {
		w.Write(http.StatusNotFound, nil, nil, middlewares.Responses.UserNotFound)
		return
	}

	if !helpers.AuthenticateHashedPassword(user.Password, opts.Password) {
		w.Write(http.StatusNotFound, nil, nil, middlewares.Responses.UserNotFound)
		return
	}

	user.Token, err = helpers.GenerateToken(user, ctx.Config.JWTSecret)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	w.WriteJSON(http.StatusOK, user, nil, "")
	return
}

func UpdateUserPassword(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

	var opts models.UpdateUserPasswordOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.UpdateUserPasswordRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.Write(http.StatusBadRequest, errs, nil, middlewares.Responses.FailedValidations)
		return
	}

	user, err := ctx.DB.GetUserByRememberToken(opts.Token)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	if user == nil {
		w.Write(http.StatusBadRequest, nil, nil, middlewares.Responses.UserNotFound)
		return
	}

	password, err := helpers.HashPassword(opts.Password)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	user.Password = password

	err = ctx.DB.UpdateUserPassword(user)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func SendRememberToken(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	w.GetRequestLanguage(r)

	var opts models.SendRememberTokenOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.SendRememberTokenRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.Write(http.StatusBadRequest, errs, nil, middlewares.Responses.FailedValidations)
		return
	}

	user, err := ctx.DB.GetUserLoginByEmail(opts.Email)
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	if user == nil {
		w.Write(http.StatusBadRequest, nil, nil, middlewares.Responses.UserNotFound)
		return
	}

	token, err := uuid.NewUUID()
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}
	err = ctx.DB.UpdateUserRememberToken(user.ID, token.String())
	if err != nil {
		w.Write(http.StatusInternalServerError, nil, err, middlewares.Responses.InternalServerError)
		return
	}

	go func(ctx *config.AppContext, user *models.User, token string) {
		ed := &helpers.EmailData{
			EmailTo:      user.Email,
			NameTo:       user.Firstname,
			EmailFrom:    ctx.Config.Mail.EmailFrom,
			NameFrom:     ctx.Config.Mail.NameFrom,
			Subject:      ctx.Config.Mail.PasswordRecover.Subject,
			TemplatePath: fmt.Sprintf("%s%s/%s", ctx.Config.Mail.Folder, ctx.Config.Mail.Path, ctx.Config.Mail.PasswordRecover.Template),
		}

		err = ed.SendEmail(models.PasswordRecoverHTML{
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			URL:       fmt.Sprintf("%s%s/%s", ctx.Config.BackofficeBaseURL, ctx.Config.BackofficePasswordRecoverPath, token),
		})
		if err != nil {
			w.LogError(err, "failed sending email")
			return
		}
		w.LogInfo(nil, "success sending email")
	}(ctx, user, token.String())

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
	return
}
