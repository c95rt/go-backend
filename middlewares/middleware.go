package middlewares

import (
	"context"
	"net/http"
	"strings"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/helpers"
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/dgrijalva/jwt-go"
	jwtmiddleware "github.com/mfuentesg/go-jwtmiddleware"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"

	"github.com/urfave/negroni"
)

func jwtErrorHandler(w http.ResponseWriter, _ *http.Request, err error) {
	r := &ResponseWriter{Writer: w}
	if err.Error() == "Token is expired" {
		r.Error(http.StatusUnauthorized, "unauthorized", WithErrorScope("token"), WithErrorType(1))
		return
	}
	if err != nil {
		r.Error(http.StatusUnauthorized, "unauthorized", WithErrorScope("token"))
	}
}

func NewJWTMiddleware(secret []byte) *jwtmiddleware.Middleware {
	return jwtmiddleware.New(
		jwtmiddleware.WithErrorHandler(jwtErrorHandler),
		jwtmiddleware.WithSigningMethod(jwt.SigningMethodHS256),
		jwtmiddleware.WithSignKey(secret),
		jwtmiddleware.WithUserProperty("_jwt-token"),
	)
}

func LoggerRequest(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestLogger := log.WithFields(log.Fields{"request_id": r.Header.Get("X-Request-ID"), "query": r.URL.Query(), "host": r.Host, "url": r.URL.Path, "headers": r.Header})
	requestLogger.Info("logger_request")
	config.SetLogger(requestLogger)
	next(rw, r)
}

func UserMiddleware() negroni.HandlerFunc {
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		authorization := r.Header.Get("Authorization")
		if len(authorization) == 0 {
			authorization = r.URL.Query().Get("token")
			r.Header.Set("Authorization", authorization)
		}
		token := strings.Split(authorization, " ")
		if len(token) == 2 {
			tokenString := token[1]
			data, _ := helpers.ParserTokenUnverified(tokenString)
			tokenParse, ok := data["u"].(map[string]interface{})
			if ok {
				id := tokenParse["i"]
				roles := tokenParse["r"]
				read := tokenParse["read"]
				email := tokenParse["email"]
				dataInfo := models.InfoUser{}
				_data := map[string]interface{}{
					"ID":    id,
					"Roles": roles,
					"Read":  read,
				}
				mapstructure.Decode(_data, &dataInfo)
				Read := dataInfo.Read
				isAdmin := helpers.Contains(dataInfo.Roles, 1)
				isCashier := helpers.Contains(dataInfo.Roles, 2)
				isReseller := helpers.Contains(dataInfo.Roles, 3)
				isClient := helpers.Contains(dataInfo.Roles, 4)
				isAPI := helpers.Contains(dataInfo.Roles, 5)
				data := map[string]interface{}{
					"Email":      email,
					"ID":         id,
					"IsAdmin":    isAdmin,
					"IsCashier":  isCashier,
					"IsReseller": isReseller,
					"IsClient":   isClient,
					"IsAPI":      isAPI,
					"Read":       Read,
					"Roles":      roles,
				}
				// ID := strconv.Itoa(dataInfo.ID)
				if r.Method != "GET" && Read {
					a := &ResponseWriter{Writer: rw}
					a.Error(http.StatusUnauthorized, "unauthorized", WithErrorScope("token"))
					return
				}
				// val, err := redisClient.Get(ID).Result()
				// if err != nil && err.Error() != "redis: nil" {
				// 	a := &ResponseWriter{Writer: rw}
				// 	a.Error(http.StatusUnauthorized, "unauthorized", WithErrorScope("token"))
				// 	return
				// }
				// if val != "1" {
				// 	a := &ResponseWriter{Writer: rw}
				// 	a.Error(http.StatusUnauthorized, "unauthorized", WithErrorScope("token"))
				// 	return
				// }
				if !isAdmin && !isCashier && !isReseller && !isClient && !isAPI && !Read {
					a := &ResponseWriter{Writer: rw}
					a.Error(http.StatusUnauthorized, "unauthorized", WithErrorScope("token"))
					return
				}
				ctx := context.WithValue(r.Context(), string("user"), data)
				next(rw, r.WithContext(ctx))
				return
			}
		}
		next(rw, r)
	})
}
