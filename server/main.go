package server

import (
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/parqueoasis/backend/config"
	"bitbucket.org/parqueoasis/backend/db"
	"bitbucket.org/parqueoasis/backend/middlewares"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joeshaw/envdecode"
	joonix "github.com/joonix/log"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/negroni"
)

func recoveryHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
			(&middlewares.ResponseWriter{Writer: w}).Error(http.StatusInternalServerError, "internal server error")
			return
		}
	}()
	next(w, r)
}

type AppHandlerFunc func(*config.AppContext, *middlewares.ResponseWriter, *http.Request)

type AppHandler struct {
	Context     *config.AppContext
	HandlerFunc AppHandlerFunc
}

func (a *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.HandlerFunc(a.Context, &middlewares.ResponseWriter{Writer: w}, r)
}

type Route struct {
	Path            string
	Handler         AppHandlerFunc
	Methods         []string
	IsProtected     bool
	IsRoleProtected bool
}

func NewRouter(ctx *config.AppContext, routes []*Route) *mux.Router {
	router := mux.NewRouter()
	for _, r := range routes {
		handler := &AppHandler{Context: ctx, HandlerFunc: r.Handler}
		if r.IsProtected {
			go router.Handle(r.Path, negroni.New(
				negroni.HandlerFunc(middlewares.NewJWTMiddleware([]byte(ctx.Config.JWTSecret)).HandlerNext),
				negroni.Wrap(handler),
			)).Methods(r.Methods...)
		}
		if !r.IsProtected && !r.IsRoleProtected {
			go router.Handle(r.Path, handler).Methods(r.Methods...)
		}
	}
	return router
}

func GetAppContext() *ContextWrapper {
	log.SetFormatter(joonix.NewFormatter())
	var conf config.Configuration
	if err := envdecode.Decode(&conf); err != nil {
		fmt.Println(fmt.Errorf("could not load the app configuration: %v", err))
		log.Fatal(err)
	}
	context := &config.AppContext{
		Config: conf,
	}

	contextWrapper := ContextWrapper{
		Context: context,
	}

	return &contextWrapper
}

type ContextWrapper struct {
	Context *config.AppContext
}

func (wrapper *ContextWrapper) CreateMySQLConnection() {
	conn, err := config.CreateConnectionSQL(wrapper.Context.Config.SQL)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetConnMaxLifetime(time.Minute * 5)
	wrapper.Context.SQLConn = conn
	wrapper.Context.DB, err = db.New(conn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("mysql: failed to connect")
	}
}

func (wrapper *ContextWrapper) CreateSMTPConnection() {
	conn := config.CreateNewConnectionSMTP(wrapper.Context.Config.AwsSMTP)
	if conn == nil {
		log.Fatal(errors.Errorf("failed connecting SMTP"))
	}
	wrapper.Context.AwsSMTP = conn
}

func (wrapper *ContextWrapper) CreateMercadoPagoIntegration() {
	mp := config.CreateMercadoPagoIntegration(wrapper.Context.Config.MercadoPago)
	if mp == nil {
		log.Fatal(errors.Errorf("failed to create mercadopago integration"))
	}
	wrapper.Context.MercadoPago = mp
}

func (wrapper *ContextWrapper) CreateNewSessionS3() {
	session, err := config.CreateNewSessionS3(wrapper.Context.Config.AwsS3)
	if err != nil {
		log.Fatal(errors.Errorf("failed to create new session s3 - %s", err.Error()))
	}
	if session == nil {
		log.Fatal(errors.Errorf("nil session s3"))
	}
	wrapper.Context.AwsS3 = session
}

func UpServer(routes []*Route, wrapper *ContextWrapper) {
	server, err := createServer(wrapper.Context, routes)
	if err != nil {
		log.Fatal(err)
	}

	if wrapper.Context.SQLConn != nil {
		defer wrapper.Context.SQLConn.Close()
	}

	log.Info("Environment " + wrapper.Context.Config.Environment)
	log.Info("Listening on " + server.Addr)

	log.Fatal(server.ListenAndServe())
}

func createServer(context *config.AppContext, routes []*Route) (*http.Server, error) {
	n := negroni.New()
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "PATCH", "HEAD"},
		AllowedHeaders: []string{"Origin", "X-Requested-With", "Content-Type", "Accept", "Authorization", "x-environment"},
	})
	n.Use(c)
	n.UseFunc(recoveryHandler)
	n.Use(negroni.HandlerFunc(middlewares.LoggerRequest))
	n.Use(middlewares.UserMiddleware())
	go n.UseHandler(NewRouter(context, routes))

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", context.Config.Port),
		ReadTimeout:  time.Duration(context.Config.Timeout) * time.Second,
		WriteTimeout: time.Duration(context.Config.Timeout) * time.Second,
		Handler:      n,
	}, nil
}
