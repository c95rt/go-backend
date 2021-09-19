package models

import (
	"time"

	"github.com/thedevsaddam/govalidator"
)

type InsertAdminUserOpts struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	DNI       string `json:"dni"`
	Phone     string `json:"phone"`
	Roles     []int  `json:"roles"`
}

var InsertAdminUserRules = govalidator.MapData{
	"email":     []string{"required", "email"},
	"password":  []string{"required"},
	"firstname": []string{"required"},
	"lastname":  []string{"required"},
	"dni":       []string{"required"},
	"phone":     []string{"required"},
	"roles":     []string{"required", "array_int"},
}

type InsertUserOpts struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	DNI       string `json:"dni"`
	Phone     string `json:"phone"`
}

var InsertUserRules = govalidator.MapData{
	"email":     []string{"required", "email"},
	"password":  []string{"required"},
	"firstname": []string{"required"},
	"lastname":  []string{"required"},
	"dni":       []string{"required"},
	"phone":     []string{"required"},
}

type GetUsersOpts struct {
	CreatedFrom string   `schema:"created_from"`
	CreatedTo   string   `schema:"created_to"`
	UserIDs     []int    `schema:"user_ids"`
	RoleIDs     []int    `schema:"role_ids"`
	Emails      []string `schema:"email"`
	Firstnames  []string `schema:"firstname"`
	Lastnames   []string `schema:"lastname"`
	Phones      []string `schema:"phone"`
	DNIs        []string `schema:"dni"`
	LimitFrom   int      `schema:"limit_from"`
	LimitTo     int      `schema:"limit_to"`
}

var GetUsersRules = govalidator.MapData{
	"created_from": []string{"date_ISO8601"},
	"created_to":   []string{"date_ISO8601"},
	"user_ids":     []string{"array_int"},
	"role_ids":     []string{"array_int"},
	"emails":       []string{"array_string"},
	"firstname":    []string{"array_string"},
	"lastname":     []string{"array_string"},
	"phones":       []string{"array_string"},
	"dnis":         []string{"array_string"},
	"limit_from":   []string{"numeric"},
	"limit_to":     []string{"numeric"},
}

type UpdateUserOpts struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	DNI       string `json:"dni"`
	Phone     string `json:"phone"`
	Roles     []int  `json:"roles"`
}

var UpdateUserRules = govalidator.MapData{
	"email":     []string{"required", "email"},
	"password":  []string{},
	"firstname": []string{"required"},
	"lastname":  []string{"required"},
	"dni":       []string{"required"},
	"phone":     []string{"required"},
	"roles":     []string{"required", "array_int"},
}

type InfoUser struct {
	ID         int
	IsAdmin    bool
	IsCashier  bool
	IsReseller bool
	IsClient   bool
	IsAPI      bool
	Read       bool
	Roles      []int
	Email      string
}

type User struct {
	ID        int    `json:"id,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Email     string `json:"email,omitempty"`
	Password  string `json:"-"`

	Created time.Time `json:"created,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
	Active  bool      `json:"active"`

	Token         string          `json:"token,omitempty"`
	RememberToken string          `json:"remember_token,omitempty"`
	Roles         []Role          `json:"role,omitempty"`
	Additional    *UserAdditional `json:"additional,omitempty"`
}

func (user *User) HasRole(roleID int) bool {
	for _, role := range user.Roles {
		if role.ID == roleID {
			return true
		}
	}
	return false
}

type Role struct {
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
type UserAdditional struct {
	ID    int    `json:"id,omitempty"`
	DNI   string `json:"dni,omitempty"`
	Phone string `json:"phone,omitempty"`
}

type UsersStruct struct {
	Users []User `json:"users"`
	Total int    `json:"total"`
}
