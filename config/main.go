package config

import (
	"fmt"
	"strconv"

	db "bitbucket.org/parqueoasis/backend/db"
	mercadopago "bitbucket.org/parqueoasis/backend/mercadopago"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gomail.v2"
)

type Configuration struct {
	JWTSecret   string `env:"JWT_SECRET,required"`
	Port        int    `env:"PORT,default=3001"`
	Timeout     int    `env:"TIMEOUT,default=1"`
	DB          db.Storage
	SQL         database
	AwsSMTP     awsSMTP
	AwsS3       awsS3
	MercadoPago mercadopagoConf
	Mail        mail
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

type awsSMTP struct {
	SMTPHost     string `env:"SMTP_HOST,required"`
	SMTPPort     int    `env:"SMTP_PORT,required"`
	SMTPUser     string `env:"SMTP_USER,required"`
	SMTPPassword string `env:"SMTP_PASSWORD,required"`
}

type mercadopagoConf struct {
	BaseURL         string `env:"MERCADOPAGO_BASEURL"`
	Token           string `env:"MERCADOPAGO_TOKEN"`
	PathPreferences string `env:"MERCADOPAGO_PATH_PREFERENCES"`
	NotificationURL string `env:"MERCADOPAGO_NOTIFICATION_URL"`
	GetPaymentURL   string `env:"MERCADOPAGO_GET_PAYMENT_URL"`
}

type awsS3 struct {
	S3Region     string `env:"S3_REGION,required"`
	S3Bucket     string `env:"S3_BUCKET,required"`
	S3Url        string `env:"S3_URL,required"`
	S3PathTicket string `env:"S3_PATH_TICKET,default=ticket"`
	S3PathOrder  string `env:"S3_PATH_ORDER,default=order"`
}

type mail struct {
	PaymentSuccess mailPaymentSuccess
	NameFrom       string `env:"MAIL_NAME_FROM"`
	EmailFrom      string `env:"MAIL_EMAIL_FROM"`
	Folder         string `env:"MAIL_FOLDER"`
	Path           string `env:"MAIL_PATH"`
}

type mailPaymentSuccess struct {
	Subject  string `env:"MAIL_PAYMENT_SUCCESS_SUBJECT"`
	Template string `env:"MAIL_PAYMENT_SUCCESS_TEMPLATE"`
	FileName string `env:"MAIL_PAYMENT_SUCCESS_FILENAME"`
}

type AppContext struct {
	Config      Configuration
	SQLConn     *sqlx.DB
	DB          db.Storage
	AwsSMTP     *gomail.Dialer
	AwsS3       *session.Session
	MercadoPago *mercadopago.MP
}

func CreateConnectionSQL(conf database) (*sqlx.DB, error) {
	conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", conf.User, conf.Password, conf.URL, strconv.Itoa(conf.Port), conf.Name)
	connection, err := sqlx.Connect("mysql", conn)
	if err != nil {
		return nil, err
	}
	return connection, nil
}

func CreateNewConnectionSMTP(conf awsSMTP) *gomail.Dialer {
	conn := gomail.NewDialer(conf.SMTPHost, conf.SMTPPort, conf.SMTPUser, conf.SMTPPassword)
	return conn
}

func CreateMercadoPagoIntegration(conf mercadopagoConf) *mercadopago.MP {
	mp := mercadopago.MP{
		BaseURL:         conf.BaseURL,
		Token:           conf.Token,
		PathPreferences: conf.PathPreferences,
		NotificationURL: conf.NotificationURL,
		GetPaymentURL:   conf.GetPaymentURL,
	}

	return &mp
}

func CreateNewSessionS3(conf awsS3) (*session.Session, error) {
	s, err := session.NewSession(&aws.Config{Region: aws.String(conf.S3Region)})
	return s, err
}

var logger *log.Entry

func SetLogger(newLogger *log.Entry) {
	logger = newLogger
}

func GetLogger() *log.Entry {
	return logger
}
