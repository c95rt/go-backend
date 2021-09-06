package models

import (
	"github.com/thedevsaddam/govalidator"
)

type LoginOpts struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserPasswordOpts struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type SendRememberTokenOpts struct {
	Email string `json:"email"`
}

var LoginRules = govalidator.MapData{
	"email":    []string{"required", "email"},
	"password": []string{"required"},
}

var UpdateUserPasswordRules = govalidator.MapData{
	"token":    []string{"required"},
	"password": []string{"required"},
}

var SendRememberTokenRules = govalidator.MapData{
	"email": []string{"required"},
}
