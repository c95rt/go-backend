package db

import (
	"database/sql"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

var ConstPaymentStatuses = struct {
	Created    models.PaymentStatus
	Rejected   models.PaymentStatus
	Approved   models.PaymentStatus
	Reversed   models.PaymentStatus
	Processing models.PaymentStatus
}{
	Created: models.PaymentStatus{
		ID:   1,
		Name: "Creado",
	},
	Rejected: models.PaymentStatus{
		ID:   2,
		Name: "Rechazado",
	},
	Approved: models.PaymentStatus{
		ID:   3,
		Name: "Aprobado",
	},
	Reversed: models.PaymentStatus{
		ID:   4,
		Name: "Reversado",
	},
	Processing: models.PaymentStatus{
		ID:   5,
		Name: "Procesando",
	},
}

var ConstPaymentMethods = struct {
	Cashier     models.PaymentMethod
	MercadoPago models.PaymentMethod
}{
	Cashier: models.PaymentMethod{
		ID:   1,
		Name: "Cajero",
	},
	MercadoPago: models.PaymentMethod{
		ID:   2,
		Name: "Mercado Pago",
	},
}

type PaymentStorage interface {
	InsertPayment(*InsertPaymentOpts) (int, error)
	GetPaymentStatusByMethodIDAndMethodStatusName(methodID int, statusName string) (*models.PaymentStatus, error)
	UpdatePaymentStatus(externalReference string, statusID int) error
}

type InsertPaymentOpts struct {
	MethodID     int    `json:"method_id"`
	Amount       int    `json:"amount"`
	UserID       int    `json:"user_id"`
	PreferenceID string `json:"preference_id"`
	OrderID      int    `json:"order_id"`
	StatusID     int    `json:"status_id"`
}

const (
	insertPayment = `
	INSERT
		payment
	SET
		method_id = :method_id,
		amount = :amount,
		user_id = :user_id,
		preference_id = :preference_id,
		order_id = :order_id,
		status_id = :status_id
	`

	getPaymentStatusByMethodIDAndMethodStatusName = `
	SELECT
		payment_status.id,
		payment_status.name
	FROM
		payment_method_status
	INNER JOIN
		payment_status ON (payment_status.id = payment_method_status.status_id)
	WHERE
		payment_method_status.method_id = :method_id AND
		payment_method_status.name = :name
	`

	updatePaymentStatus = `
	UPDATE
		payment
	SET
		status_id = :status_id,
		updated = current_timestamp()
	WHERE
		preference_id = :external_reference
	`
)

func (db *DB) InsertPayment(opts *InsertPaymentOpts) (int, error) {
	tx, err := db.NewTx()
	if err != nil {
		return 0, errors.Wrap(err, "failed to start transaction")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	id, newErr := db.insertPaymentTx(tx, opts)
	if newErr != nil {
		err = newErr
		return 0, err
	}

	return id, nil
}

func (db *DB) insertPaymentTx(tx Tx, opts *InsertPaymentOpts) (int, error) {
	stmt, err := tx.PrepareNamed(insertPayment)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
		"method_id":     opts.MethodID,
		"amount":        opts.Amount,
		"user_id":       opts.UserID,
		"preference_id": opts.PreferenceID,
		"order_id":      opts.OrderID,
		"status_id":     opts.StatusID,
	}

	result, err := stmt.Exec(args)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if int(rowsAffected) != 1 {
		return 0, errors.Errorf("expected %d and inserted %d", 1, rowsAffected)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (db *DB) GetPaymentStatusByMethodIDAndMethodStatusName(methodID int, statusName string) (*models.PaymentStatus, error) {
	stmt, err := db.PrepareNamed(getPaymentStatusByMethodIDAndMethodStatusName)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"method_id": methodID,
		"name":      statusName,
	}

	var status models.PaymentStatus

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&status.ID,
		&status.Name,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &status, nil
}

func (db *DB) UpdatePaymentStatus(externalReference string, statusID int) error {
	tx, err := db.NewTx()
	if err != nil {
		return errors.Wrap(err, "failed to start transaction")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	err = db.updatePaymentStatusTx(tx, externalReference, statusID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updatePaymentStatusTx(tx Tx, externalReference string, statusID int) error {
	stmt, err := tx.PrepareNamed(updatePaymentStatus)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"status_id":          statusID,
		"external_reference": externalReference,
	}

	result, err := stmt.Exec(args)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if int(rowsAffected) != 1 {
		return errors.Errorf("expected %d and updated %d", 1, rowsAffected)
	}

	return nil
}
