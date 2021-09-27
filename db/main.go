package db

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const isoLayout = "2006-01-02"
const maxRetries = 3

type Storage interface {
	UserStorage
	AuthStorage
	EventStorage
	OrderStorage
	PaymentStorage
	TicketStorage
}

type db interface {
	NewTx() (Tx, error)
}

type conn interface {
	Rebind(string) string
	NamedExec(string, interface{}) (sql.Result, error)
	Select(interface{}, string, ...interface{}) error
	QueryRow(string, ...interface{}) *sql.Row
	PrepareNamed(string) (*sqlx.NamedStmt, error)
	Get(interface{}, string, ...interface{}) error
	Query(string, ...interface{}) (*sql.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Preparex(query string) (*sqlx.Stmt, error)
	Exec(string, ...interface{}) (sql.Result, error)
}

type Tx interface {
	conn

	Commit() error
	Rollback() error
}

type transactorImpl struct {
	*sqlx.DB
}

func (t *transactorImpl) NewTx() (Tx, error) {
	return t.Beginx()
}

type DB struct {
	conn
	db
}

func New(db *sqlx.DB) (*DB, error) {
	var (
		dbWrapper *DB
		err       error
	)

	tries := maxRetries
	for tries >= 0 {
		time.Sleep(1 * time.Second)

		log.WithFields(log.Fields{
			"retries_left": tries,
		}).Warnf("%s: trying to connect to create connection", db.DriverName())

		dbWrapper, err = tryOpenConnection(db)
		if err != nil {
			if tries == 0 {
				return nil, err
			}

			tries = tries - 1
			continue
		}

		break
	}

	return dbWrapper, nil
}

func tryOpenConnection(db *sqlx.DB) (*DB, error) {
	// TODO: move the db connection code to
	// this func

	err := db.Ping()
	if err != nil {
		return nil, errors.Wrap(err, "failed to ping db")
	}

	return &DB{
		db,
		&transactorImpl{db},
	}, nil
}
