package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

type OrderStorage interface {
	InsertOrder(userID int, clientID int, eventID int, tickets int) (*models.Order, error)
	GetOrderByID(orderID int) (*models.Order, error)
	GetOrderByExternalReference(externalReference string) (*models.Order, error)
	GetOrderByTransactionID(transactionID string) (*models.Order, error)
	UpdateOrder(orderID int, eventID int) error
	UseOrder(orderID int, userID int) error
	GetOrders(opts *models.GetOrdersOpts) (*models.GetOrdersStruct, error)
	GetSalesSummary() ([]models.DailySales, error)
	GetCashierSummary(cashierID int, dateFrom string, dateTo string) ([]models.CashierMonthlySales, error)
}

const (
	insertOrder = `
	INSERT
		orders
	SET
		user_id = :user_id,
		client_id = :client_id,
		transaction_id = :transaction_id,
		event_id = :event_id,
		tickets = :tickets,
		initial_tickets = :initial_tickets
	`

	getOrderByTransactionID = `
	SELECT
		orders.id,
		orders.transaction_id,
		orders.tickets,
		orders.initial_tickets,
		orders.created,
		orders.updated,
		event.id,
		event.price,
		event.start_date_time,
		event.end_date_time
		COALESCE(
			(
				SELECT
					JSON_OBJECT(
						'id', payment.id,
						'amount', payment.amount,
						'reference_id', payment.preference_id,
						'created', DATE_FORMAT(payment.created, :iso8601),
						'updated', DATE_FORMAT(payment.updated, :iso8601),
						'user', JSON_OBJECT(
							'id', payment_user.id,
							'firstname', payment_user.firstname,
							'lastname', payment_user.lastname,
							'email', payment_user.email
						),
						'status', JSON_OBJECT(
							'id', payment_status.id,
							'name', payment_status.name
						),
						'method', JSON_OBJECT(
							'id', payment_method.id,
							'name', payment_method.name
						)
					)
				FROM
					payment
				LEFT JOIN
					payment_method ON (payment_method.id = payment.method_id)
				LEFT JOIN
					payment_status ON (payment_status.id = payment.status_id)
				LEFT JOIN
					user payment_user ON (payment.user_id = payment_user.id)
				WHERE
					payment.id = (
						SELECT
							id
						FROM
							payment
						WHERE
							order_id = orders.id
						ORDER BY
							id DESC
						LIMIT 1
					) AND
					payment.active = true
			), '{}'
		)
	FROM
		orders
	INNER JOIN
		event ON (event.id = orders.event_id AND event.active = true)
	WHERE
		orders.active = true AND
		orders.transaction_id = :transaction_id
	`

	getOrderByExternalReference = `
	SELECT
		orders.id,
		orders.transaction_id,
		orders.tickets,
		orders.initial_tickets,
		orders.created,
		orders.updated,
		user.id,
		user.firstname,
		user.lastname,
		user.email,
		client.id,
		client.firstname,
		client.lastname,
		client.email,
		event.id,
		event.price,
		event.start_date_time,
		event.end_date_time,
		event_type.id,
		event_type.name,
		COALESCE(
			(
				SELECT
					JSON_OBJECT(
						'id', payment.id,
						'amount', payment.amount,
						'reference_id', payment.preference_id,
						'created', DATE_FORMAT(payment.created, :iso8601),
						'updated', DATE_FORMAT(payment.updated, :iso8601),
						'user', JSON_OBJECT(
							'id', payment_user.id,
							'firstname', payment_user.firstname,
							'lastname', payment_user.lastname,
							'email', payment_user.email
						),
						'status', JSON_OBJECT(
							'id', payment_status.id,
							'name', payment_status.name
						),
						'method', JSON_OBJECT(
							'id', payment_method.id,
							'name', payment_method.name
						)
					)
				FROM
					payment
				LEFT JOIN
					payment_method ON (payment_method.id = payment.method_id)
				LEFT JOIN
					payment_status ON (payment_status.id = payment.status_id)
				LEFT JOIN
					user payment_user ON (payment.user_id = payment_user.id)
				WHERE
					payment.id = (
						SELECT
							id
						FROM
							payment
						WHERE
							order_id = orders.id
						ORDER BY
							id DESC
						LIMIT 1
					) AND
					payment.active = true
			), '{}'
		)
	FROM
		orders
	INNER JOIN
		event ON (event.id = orders.event_id AND event.active = true)
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	INNER JOIN
		payment ON (payment.order_id = orders.id AND payment.preference_id = :external_reference)
	INNER JOIN
		user ON (user.id = orders.user_id)
	INNER JOIN
		user AS client ON (client.id = orders.client_id)
	WHERE
		orders.active = true
	GROUP BY
		orders.id
	`

	getOrderByID = `
	SELECT
		orders.id,
		orders.transaction_id,
		orders.tickets,
		orders.initial_tickets,
		orders.created,
		orders.updated,
		user.id,
		user.firstname,
		user.lastname,
		user.email,
		client.id,
		client.firstname,
		client.lastname,
		client.email,
		event.id,
		event.price,
		event.start_date_time,
		event.end_date_time,
		event_type.id,
		event_type.name,
		COALESCE(
			(
				SELECT
					JSON_OBJECT(
						'id', payment.id,
						'amount', payment.amount,
						'reference_id', payment.preference_id,
						'created', DATE_FORMAT(payment.created, :iso8601),
						'updated', DATE_FORMAT(payment.updated, :iso8601),
						'user', JSON_OBJECT(
							'id', payment_user.id,
							'firstname', payment_user.firstname,
							'lastname', payment_user.lastname,
							'email', payment_user.email
						),
						'status', JSON_OBJECT(
							'id', payment_status.id,
							'name', payment_status.name
						),
						'method', JSON_OBJECT(
							'id', payment_method.id,
							'name', payment_method.name
						)
					)
				FROM
					payment
				LEFT JOIN
					payment_method ON (payment_method.id = payment.method_id)
				LEFT JOIN
					payment_status ON (payment_status.id = payment.status_id)
				LEFT JOIN
					user payment_user ON (payment.user_id = payment_user.id)
				WHERE
					payment.id = (
						SELECT
							id
						FROM
							payment
						WHERE
							order_id = orders.id
						ORDER BY
							id DESC
						LIMIT 1
					) AND
					payment.active = true
			), '{}'
		)
	FROM
		orders
	INNER JOIN
		user ON (user.id = orders.user_id)
	INNER JOIN
		user AS client ON (client.id = orders.client_id)
	INNER JOIN
		event ON (event.id = orders.event_id AND event.active = true)
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	WHERE
		orders.active = true AND
		orders.id = :id
	`

	updateOrder = `
	UPDATE
		orders
	SET
		event_id = :event_id,
		updated = current_timestamp()
	WHERE
		id = :order_id AND
		active = true AND
		tickets > 0;
	`

	useOrder = `
	UPDATE
		orders
	SET
		tickets = tickets-1,
		updated = current_timestamp();
	WHERE
		orders.id = :order_id AND
		orders.tickets > 0 AND
		orders.active = true;
	`

	getOrders = `
	SELECT
		orders.id,
		orders.transaction_id,
		orders.tickets,
		orders.initial_tickets,
		orders.created,
		orders.updated,
		client.id,
		client.firstname,
		client.lastname,
		user.id,
		user.firstname,
		user.lastname,
		IF(payment.id IS NOT NULL, true, false) AS paid,
		event.id,
		event.start_date_time,
		event.end_date_time,
		event.price,
		event_type.id,
		event_type.name
	FROM
		orders
	LEFT JOIN
		payment ON (payment.id = (SELECT id FROM payment WHERE order_id = orders.id AND status_id = :status_id ORDER BY id DESC LIMIT 1))
	INNER JOIN
		event ON (event.id = orders.event_id)
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	INNER JOIN
		user ON (user.id = orders.user_id)
	INNER JOIN
		user AS client ON (orders.client_id = client.id)
	WHERE
		orders.active = true
		#FILTERS#
	ORDER BY
		event.start_date_time DESC
	LIMIT :limit_to OFFSET :limit_from
	`

	countOrders = `
	SELECT
		COUNT(orders.id)
	FROM
		orders
	INNER JOIN
		event ON (event.id = orders.event_id)
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	INNER JOIN
		user ON (orders.user_id = user.id)
	INNER JOIN
		user AS client ON (orders.client_id = client.id)
	WHERE
		orders.active = true
		#FILTERS#
	`

	getSalesSummary = `
	SELECT
		orders.created,
		SUM(COALESCE(payment.amount, event.price*IF(orders.initial_tickets = 0, 1, orders.initial_tickets)))
	FROM
		orders
	INNER JOIN
		event ON (event.id = orders.event_id)
	LEFT JOIN
		payment ON (payment.id = (SELECT id FROM payment WHERE order_id = orders.id AND status_id = 3 ORDER BY id DESC LIMIT 1))
	WHERE
		YEAR(orders.created) BETWEEN YEAR(current_timestamp())-1 AND YEAR(current_timestamp())	
	GROUP BY
		YEAR(orders.created),
		MONTH(orders.created),
		DAY(orders.created)
	`

	insertOrderUse = `
	INSERT
		order_use
	SET
		user_id = :user_id,
		order_id = :order_id
	`

	getCashierSummary = `
	SELECT
		user.id,
		user.firstname,
		user.lastname,
		user.email,
		orders.created,
		SUM(IF(COALESCE((SELECT true FROM payment WHERE payment.order_id = orders.id AND payment.status_id = :status_id ORDER BY payment.id DESC LIMIT 1), false), orders.initial_tickets, 0)),
		(
			SELECT
				COUNT(order_use.id)
			FROM
				order_use
			WHERE
				YEAR(order_use.created) = YEAR(orders.created) AND
				MONTH(order_use.created) = MONTH(orders.created) AND
				order_use.user_id = user.id
		) AS uses
	FROM
		orders
	INNER JOIN
		user ON (user.id = orders.user_id)
	WHERE
		orders.created BETWEEN :date_from AND :date_to AND
		orders.user_id = :cashier_id
	GROUP BY
		YEAR(orders.created),
		MONTH(orders.created)
	`
)

func (db *DB) InsertOrder(userID int, clientID int, eventID int, tickets int) (*models.Order, error) {
	tx, err := db.NewTx()
	if err != nil {
		return nil, errors.Wrap(err, "failed to start transaction")
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
	}()

	transactionID := GenerateTicketUUID()

	orderID, newErr := db.insertOrderTx(tx, userID, clientID, eventID, transactionID, tickets)
	if newErr != nil {
		err = newErr
		return nil, err
	}

	order := models.Order{
		ID: orderID,
		User: &models.User{
			ID: userID,
		},
		Client: &models.User{
			ID: clientID,
		},
		Tickets:       tickets,
		TransactionID: transactionID,
	}

	return &order, nil
}

func (db *DB) insertOrderTx(tx Tx, userID int, clientID int, eventID int, transactionID string, tickets int) (int, error) {
	stmt, err := tx.PrepareNamed(insertOrder)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
		"event_id":        eventID,
		"user_id":         userID,
		"client_id":       clientID,
		"tickets":         tickets,
		"initial_tickets": tickets,
		"transaction_id":  transactionID,
	}

	result, err := stmt.Exec(args)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (db *DB) GetOrderByExternalReference(externalReference string) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByExternalReference)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"external_reference": externalReference,
		"iso8601":            ConstIso8061,
	}

	var order models.Order
	var user models.User
	var client models.User
	var event models.Event
	var eventType models.EventType
	var paymentBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&order.ID,
		&order.TransactionID,
		&order.Tickets,
		&order.InitialTickets,
		&order.Created,
		&order.Updated,
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&client.ID,
		&client.Firstname,
		&client.Lastname,
		&client.Email,
		&event.ID,
		&event.Price,
		&event.StartDateTime,
		&event.EndDateTime,
		&eventType.ID,
		&eventType.Name,
		&paymentBT,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(paymentBT, &order.Payment); err != nil {
		return nil, err
	}

	order.User = &user
	order.Client = &client
	event.Type = &eventType
	order.Event = &event
	order.Price = event.Price * order.InitialTickets

	return &order, nil
}

func (db *DB) GetOrderByTransactionID(transactionID string) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByTransactionID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"transaction_id": transactionID,
		"iso8601":        ConstIso8061,
	}

	var event models.Event
	var order models.Order
	var paymentBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&order.ID,
		&order.TransactionID,
		&order.Tickets,
		&order.InitialTickets,
		&order.Created,
		&order.Updated,
		&event.ID,
		&event.Price,
		&event.StartDateTime,
		&event.EndDateTime,
		&paymentBT,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(paymentBT, &order.Payment); err != nil {
		return nil, err
	}

	order.Event = &event
	order.Price = event.Price * order.InitialTickets

	return &order, nil
}

func (db *DB) GetOrderByID(id int) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"id":      id,
		"iso8601": ConstIso8061,
	}

	var event models.Event
	var eventType models.EventType
	var order models.Order
	var paymentBT []byte
	var user models.User
	var client models.User

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&order.ID,
		&order.TransactionID,
		&order.Tickets,
		&order.InitialTickets,
		&order.Created,
		&order.Updated,
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&client.ID,
		&client.Firstname,
		&client.Lastname,
		&client.Email,
		&event.ID,
		&event.Price,
		&event.StartDateTime,
		&event.EndDateTime,
		&eventType.ID,
		&eventType.Name,
		&paymentBT,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(paymentBT, &order.Payment); err != nil {
		return nil, err
	}

	event.Type = &eventType
	order.Event = &event
	order.User = &user
	order.Client = &client
	order.Price = event.Price * order.InitialTickets

	return &order, nil
}

func (db *DB) UpdateOrder(orderID int, eventID int) error {
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

	err = db.updateOrderTx(tx, orderID, eventID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updateOrderTx(tx Tx, orderID int, eventID int) error {
	stmt, err := tx.PrepareNamed(updateOrder)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"order_id": orderID,
		"event_id": eventID,
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

func (db *DB) UseOrder(orderID int, userID int) error {
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

	err = db.useOrderTx(tx, orderID)
	if err != nil {
		return err
	}

	err = db.insertOrderUseTx(tx, orderID, userID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) useOrderTx(tx Tx, orderID int) error {
	stmt, err := tx.PrepareNamed(useOrder)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"order_id": orderID,
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
		return errors.Errorf("expected %d and deleted %d", 1, rowsAffected)
	}

	return nil
}

func (db *DB) insertOrderUseTx(tx Tx, orderID int, userID int) error {
	stmt, err := tx.PrepareNamed(insertOrderUse)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"user_id":  userID,
		"order_id": orderID,
	}

	_, err = stmt.Exec(args)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) GetOrders(opts *models.GetOrdersOpts) (*models.GetOrdersStruct, error) {
	var filters string
	args := make(map[string]interface{})
	if opts.EventFrom != "" {
		filters += " AND event.start_date_time >= :event_from "
		args["event_from"] = opts.EventFrom
	}
	if opts.EventTo != "" {
		filters += " AND event.start_date_time <= :event_to "
		args["event_to"] = opts.EventFrom
	}
	if opts.TransactionID != "" {
		filters += " AND orders.transaction_id <= :transaction_id "
		args["transaction_id"] = opts.TransactionID
	}
	if opts.EventTypeID != 0 {
		filters += " AND event.event_type_id = :event_type_id "
	}
	if opts.Paid != nil {
		filters += " AND COALESCE((SELECT true FROM payment WHERE payment.order_id = orders.id AND payment.status_id = :status_id ORDER BY payment.id DESC LIMIT 1), false) = :paid"
		args["paid"] = opts.Paid
		fmt.Println(filters)
	}
	if opts.LimitTo == 0 {
		opts.LimitTo = 10
	}
	args["limit_to"] = opts.LimitTo
	args["limit_from"] = opts.LimitFrom
	args["status_id"] = ConstPaymentStatuses.Approved.ID

	totalOrders, err := db.countOrders(filters, args)
	if err != nil {
		return nil, err
	}

	query := strings.ReplaceAll(getOrders, "#FILTERS#", filters)

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	orders := models.GetOrdersStruct{
		Total: totalOrders,
	}

	for rows.Next() {
		var order models.Order
		var client models.User
		var user models.User
		var event models.Event
		var eventType models.EventType
		if err := rows.Scan(
			&order.ID,
			&order.TransactionID,
			&order.Tickets,
			&order.InitialTickets,
			&order.Created,
			&order.Updated,
			&client.ID,
			&client.Firstname,
			&client.Lastname,
			&user.ID,
			&user.Firstname,
			&user.Lastname,
			&order.Paid,
			&event.ID,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Price,
			&eventType.ID,
			&eventType.Name,
		); err != nil {
			return nil, err
		}

		order.Client = &client
		order.User = &user
		order.Event = &event
		order.Price = event.Price * order.InitialTickets

		orders.Orders = append(orders.Orders, order)
	}

	return &orders, nil
}

func (db *DB) countOrders(filters string, args map[string]interface{}) (int, error) {
	query := strings.ReplaceAll(countOrders, "#FILTERS#", filters)
	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return 0, err
	}

	row := stmt.QueryRow(args)
	var total int
	if err := row.Scan(
		&total,
	); err != nil {
		return 0, err
	}

	return total, nil
}

func (db *DB) GetSalesSummary() ([]models.DailySales, error) {
	rows, err := db.Query(getSalesSummary)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var dailySalesArray []models.DailySales
	for rows.Next() {
		var dailySales models.DailySales
		if err := rows.Scan(
			&dailySales.Date,
			&dailySales.Total,
		); err != nil {
			return nil, err
		}
		dailySalesArray = append(dailySalesArray, dailySales)
	}

	return dailySalesArray, nil
}

func (db *DB) GetCashierSummary(cashierID int, dateFrom string, dateTo string) ([]models.CashierMonthlySales, error) {
	stmt, err := db.PrepareNamed(getCashierSummary)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"cashier_id": cashierID,
		"date_from":  dateFrom,
		"date_to":    dateTo,
		"status_id":  ConstPaymentStatuses.Approved.ID,
	}

	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var cashierMonthlySales []models.CashierMonthlySales
	for rows.Next() {
		var user models.User
		var sales models.CashierMonthlySales
		if err := rows.Scan(
			&user.ID,
			&user.Firstname,
			&user.Lastname,
			&user.Email,
			&sales.Date,
			&sales.TotalSales,
			&sales.TotalUses,
		); err != nil {
			return nil, err
		}
		sales.User = &user
		cashierMonthlySales = append(cashierMonthlySales, sales)
	}

	return cashierMonthlySales, nil
}
