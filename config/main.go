package config

import (
	"fmt"
	"strconv"

	db "bitbucket.org/parqueoasis/backend/db"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	JWTSecret   string `env:"JWT_SECRET,required"`
	Port        int    `env:"PORT,default=3001"`
	Timeout     int    `env:"TIMEOUT,default=1"`
	DB          db.Storage
	SQL         database
	Environment string `env:"ENVIRONMENT,default=development"`
	CasbinModel string `env:"RBAC_FILE,default=config/rbac.conf"`
	AppName     string `env:"APP_NAME,default=app"`
}

type database struct {
	URL            string `env:"DATA_BASE_URL,required"`
	Name           string `env:"DATA_BASE_NAME,required"`
	User           string `env:"DATA_BASE_USER,required"`
	Port           int    `env:"DATA_BASE_PORT,default=3306"`
	Password       string `env:"DATA_BASE_PASSWORD,required"`
	OpenConnection int    `env:"DATA_BASE_MAX_OPEN_CONNECTION,default=5"`
}

type AppContext struct {
	Config  Configuration
	SQLConn *sqlx.DB
	DB      db.Storage
}

func CreateConnectionSQL(conf database) (*sqlx.DB, error) {
	conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", conf.User, conf.Password, conf.URL, strconv.Itoa(conf.Port), conf.Name)
	connection, err := sqlx.Connect("mysql", conn)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

var logger *log.Entry

func SetLogger(newLogger *log.Entry) {
	logger = newLogger
}

func GetLogger() *log.Entry {
	return logger
}
