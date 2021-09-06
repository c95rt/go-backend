package helpers

import (
	"fmt"
	"reflect"
	"time"

	"github.com/thedevsaddam/govalidator"
)

func init() {
	govalidator.AddCustomRule("array_int", func(field string, rule string, message string, value interface{}) error {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Array || rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice {
			arr := value.([]int)
			for _, v := range arr {
				if v <= 0 {
					if message != "" {
						return fmt.Errorf(message)
					}
					return fmt.Errorf("The %s field must be array of int higher 0", field)
				}
			}
		}
		return nil
	})
	govalidator.AddCustomRule("date_ISO8601", func(field string, rule string, message string, value interface{}) error {
		dateLayoutISO8601 := "2006-01-02"
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.String {
			date := value.(string)
			if _, err := time.Parse(dateLayoutISO8601, date); err != nil {
				if message != "" {
					return fmt.Errorf(message)
				}
				return fmt.Errorf("The %s field must be ISO8601 yyyy-mm-dd date ", field)
			}
		}
		return nil
	})
	govalidator.AddCustomRule("datetime_RFC3339", func(field string, rule string, message string, value interface{}) error {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.String {
			date := value.(string)
			if _, err := time.Parse(time.RFC3339, date); err != nil {
				if message != "" {
					return fmt.Errorf(message)
				}
				return fmt.Errorf("The %s field must be RFC3339 YYYY-MM-DDTHH:mm:ssZ date time ", field)
			}
		}
		return nil
	})
	govalidator.AddCustomRule("date_time_ISO8601", func(field string, rule string, message string, value interface{}) error {
		dateTimeLayoutISO8601 := "2006-01-02 15:04:05"
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.String {
			date := value.(string)
			if _, err := time.Parse(dateTimeLayoutISO8601, date); err != nil {
				if message != "" {
					return fmt.Errorf(message)
				}
				return fmt.Errorf("The %s field must be ISO8601 yyyy-mm-dd hh:mm:ss date ", field)
			}
		}
		return nil
	})
	govalidator.AddCustomRule("time_ISO8601", func(field string, rule string, message string, value interface{}) error {
		dateTimeLayoutISO8601 := "15:04:05"
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.String {
			date := value.(string)
			if _, err := time.Parse(dateTimeLayoutISO8601, date); err != nil {
				return fmt.Errorf("The %s field must be ISO8601 hh:mm:ss time ", field)

			}
		}
		return nil
	})
	govalidator.AddCustomRule("array_string", func(field string, rule string, message string, value interface{}) error {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Array || rv.Kind() == reflect.Map || rv.Kind() == reflect.Slice {
			arr := value.([]string)
			for _, v := range arr {
				if v == "" {
					if message != "" {
						return fmt.Errorf(message)
					}
					return fmt.Errorf("The %s field must be array of string not empty", field)
				}
			}
		}
		return nil
	})
	govalidator.AddCustomRule("hour_ISO8601", func(field string, rule string, message string, value interface{}) error {
		dateTimeLayoutISO8601 := "15:04"
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.String {
			date := value.(string)
			if _, err := time.Parse(dateTimeLayoutISO8601, date); err != nil {
				return fmt.Errorf("The %s field must be ISO8601 hh:mm time ", field)

			}
		}
		return nil
	})
}
