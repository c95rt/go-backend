package api

import (
	"net/http"
	"strconv"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/db"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/middlewares"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/gorilla/mux"
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

	var opts models.InsertAdminUserOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.InsertAdminUserRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	emailCounter, dniCounter, err := ctx.DB.ValidateUserEmailAndDNI(opts.Email, opts.DNI)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed validating email and dni")
		return
	}

	if emailCounter > 0 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "email exists")
		return
	}

	if dniCounter > 0 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "dni exists")
		return
	}

	newPassword, err := helpers.HashPassword(opts.Password)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed hashing password")
		return
	}
	opts.Password = newPassword

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

func UpdateUser(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusBadRequest, nil, err, "failed parsing user id")
		return
	}

	if !(userInfo.IsAdmin && userInfo.IsCashier) && userInfo.ID != userID {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	var opts models.UpdateUserOpts
	validatorOpts := govalidator.Options{
		Request: r,
		Rules:   models.UpdateUserRules,
		Data:    &opts,
	}
	v := govalidator.New(validatorOpts)
	errs := v.ValidateJSON()
	if len(errs) > 0 {
		w.WriteJSON(http.StatusBadRequest, errs, nil, "failed validations")
		return
	}

	user, err := ctx.DB.GetUserByID(userID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting user")
		return
	}

	if user == nil {
		w.WriteJSON(http.StatusNotFound, nil, nil, "user not found")
		return
	}

	if user.Email != opts.Email || user.Additional.DNI != opts.DNI {
		emailCounter, dniCounter, err := ctx.DB.ValidateUserEmailAndDNI(opts.Email, opts.DNI)
		if err != nil {
			w.WriteJSON(http.StatusInternalServerError, nil, err, "failed validating email and dni")
			return
		}

		if user.Email != opts.Email {
			if emailCounter > 0 {
				w.WriteJSON(http.StatusBadRequest, nil, nil, "email exists")
				return
			}
		}

		if user.Additional.DNI != opts.DNI {
			if dniCounter > 0 {
				w.WriteJSON(http.StatusBadRequest, nil, nil, "dni exists")
				return
			}
		}
	}

	if opts.Password != "" {
		newPassword, err := helpers.HashPassword(opts.Password)
		if err != nil {
			w.WriteJSON(http.StatusInternalServerError, nil, err, "failed hashing password")
		}
		opts.Password = newPassword
	}

	if !(userInfo.IsAdmin || userInfo.IsCashier) {
		var newRoles []int
		for _, role := range user.Roles {
			newRoles = append(newRoles, role.ID)
		}
		opts.Roles = newRoles
		opts.Password = user.Password
	}

	if err := ctx.DB.UpdateUser(userID, &opts); err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed updating user")
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}

func GetUser(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
	userInfo := models.InfoUser{}
	mapstructure.Decode(r.Context().Value("user"), &userInfo)

	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		w.WriteJSON(http.StatusBadRequest, nil, err, "failed parsing user id")
		return
	}

	if !(userInfo.IsAdmin && userInfo.IsCashier) && userInfo.ID != userID {
		w.WriteJSON(http.StatusForbidden, nil, nil, "invalid roles")
		return
	}

	user, err := ctx.DB.GetUserByID(userID)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed getting user")
		return
	}

	w.WriteJSON(http.StatusOK, user, nil, "")
}

func InsertUser(ctx *config.AppContext, w *middlewares.ResponseWriter, r *http.Request) {
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

	emailCounter, dniCounter, err := ctx.DB.ValidateUserEmailAndDNI(opts.Email, opts.DNI)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed validating email and dni")
		return
	}

	if emailCounter > 0 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "email exists")
		return
	}

	if dniCounter > 0 {
		w.WriteJSON(http.StatusBadRequest, nil, nil, "dni exists")
		return
	}

	newPassword, err := helpers.HashPassword(opts.Password)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, err, "failed hashing password")
		return
	}
	opts.Password = newPassword

	finalOpts := models.InsertAdminUserOpts{
		Email:     opts.Email,
		Password:  opts.Password,
		Firstname: opts.Firstname,
		Lastname:  opts.Lastname,
		DNI:       opts.DNI,
		Phone:     opts.Phone,
		Roles:     []int{db.ConstRoles.Client},
	}

	_, err = ctx.DB.InsertUser(&finalOpts)
	if err != nil {
		w.WriteJSON(http.StatusInternalServerError, nil, nil, "failed inserting user")
		return
	}

	w.WriteJSON(http.StatusNoContent, nil, nil, "")
}
