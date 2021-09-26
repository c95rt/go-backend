package db

import (
	"database/sql"
	"encoding/json"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

type OrderStorage interface {
	InsertOrder(userID int, clientID int, events []models.Event, ticketsByEvent map[int]int) (*models.Order, error)
	GetOrderByTicketUUID(ticketUUID string) (*models.Order, error)
	DeleteTicket(ticketID int) error
	GetOrderByID(orderID int) (*models.Order, error)
	GetOrderByExternalReference(externalReference string) (*models.Order, error)
}

const (
	insertOrder = `
	INSERT
		orders
	SET
		user_id = :user_id,
		client_id = :client_id
	`

	insertTicket = `
	INSERT
		ticket
	SET
		uuid = :uuid,
		order_id = :order_id,
		event_id = :event_id
	`

	getOrderByTicketUUID = `
	SELECT
		ticket.id,
		event.id,
		event.start_date_time,
		event.end_date_time,
		orders.id,
		orders.created,
		orders.updated,
		COALESCE(
			(
				SELECT
					JSON_OBJECT(
						'id', payment.id,
						'amount', payment.amount,
						'reference_id', payment.preference_id,
						'created', payment.created,
						'updated', payment.updated,
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
		ticket
	INNER JOIN
		orders ON (orders.id = ticket.order_id AND orders.active = true)
		event ON (event.id = ticket.event_id AND event.active = true)
	WHERE
		ticket.uuid = :uuid AND
		ticket.active = true
	`

	getOrderByID = `
	SELECT
		orders.id,
		orders.created,
		orders.updated,
		user.id,
		user.firstname,
		user.lastname,
		user.email,
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
		),
		(
			SELECT
				COALESCE(
					CONCAT('[',
						GROUP_CONCAT(
							JSON_EXTRACT(JSON_OBJECT(
								'id', ticket.id,
								'uuid', ticket.uuid,
								'event', JSON_OBJECT(
									'id', event.id,
									'price', event.price,
									'start_date_time', DATE_FORMAT(event.start_date_time, :iso8601),
									'end_date_time', DATE_FORMAT(event.end_date_time, :iso8601)
								)
							), '$')
						),
					']'),
				'[]')
			FROM
				ticket
			INNER JOIN
				event ON (event.id = ticket.event_id AND event.active = true)
			WHERE
				ticket.order_id = orders.id AND
				ticket.active = true
			GROUP BY
				ticket.order_id
		)
	FROM
		orders
	INNER JOIN
		user ON (user.id = orders.client_id)
	WHERE
		orders.id = :order_id AND
		orders.active = true
	GROUP BY
		orders.id
	`
	getOrderByExternalReference = `
	SELECT
		orders.id,
		orders.created,
		orders.updated,
		user.id,
		user.firstname,
		user.lastname,
		user.email,
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
		),
		(
			SELECT
				COALESCE(
					CONCAT('[',
						GROUP_CONCAT(
							JSON_EXTRACT(JSON_OBJECT(
								'id', ticket.id,
								'uuid', ticket.uuid,
								'event', JSON_OBJECT(
									'id', event.id,
									'price', event.price,
									'start_date_time', DATE_FORMAT(event.start_date_time, :iso8601),
									'end_date_time', DATE_FORMAT(event.end_date_time, :iso8601)
								)
							), '$')
						),
					']'),
				'[]')
			FROM
				ticket
			INNER JOIN
				event ON (event.id = ticket.event_id AND event.active = true)
			WHERE
				ticket.order_id = orders.id AND
				ticket.active = true
			GROUP BY
				ticket.order_id
		)
	FROM
		orders
	INNER JOIN
		payment ON (payment.order_id = orders.id AND payment.preference_id = :external_reference)
	INNER JOIN
		user ON (user.id = orders.client_id)
	WHERE
		orders.active = true
	GROUP BY
		orders.id
	`
)

func (db *DB) InsertOrder(userID int, clientID int, events []models.Event, ticketsByEvent map[int]int) (*models.Order, error) {
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

	orderID, err := db.insertOrderTx(tx, userID, clientID)
	if err != nil {
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
	}

	for _, event := range events {
		for i := 0; i < ticketsByEvent[event.ID]; i++ {
			ticket := models.Ticket{
				Event: &event,
			}
			ticket.ID, err = db.insertTicketTx(tx, orderID, event.ID, i)
			if err != nil {
				return nil, err
			}
			order.Tickets = append(order.Tickets, ticket)
			order.Price += event.Price
		}
	}

	return &order, nil
}

func (db *DB) insertOrderTx(tx Tx, userID int, clientID int) (int, error) {
	stmt, err := tx.PrepareNamed(insertOrder)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
		"user_id":   userID,
		"client_id": clientID,
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

func (db *DB) insertTicketTx(tx Tx, orderID int, eventID int, time int) (int, error) {
	stmt, err := tx.PrepareNamed(insertTicket)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
		"uuid":     GenerateTicketUUID(orderID, time),
		"order_id": orderID,
		"event_id": eventID,
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

func (db *DB) GetOrderByTicketUUID(ticketUUID string) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByTicketUUID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"uuid": ticketUUID,
	}

	var ticket models.Ticket
	var event models.Event
	var order models.Order
	var paymentBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&ticket.ID,
		&event.ID,
		&event.StartDateTime,
		&event.EndDateTime,
		&order.ID,
		&order.Created,
		&order.Updated,
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

	ticket.Event = &event
	order.Tickets = append(order.Tickets, ticket)

	return &order, nil
}

func (db *DB) DeleteTicket(ticketID int) error {
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

	err = db.deleteTicketTx(tx, ticketID)
	if err != nil {
		return err
	}

	return nil
}

const (
	deleteTicket = `
	UPDATE
		ticket
	SET
		active = :active
	WHERE
		ticket.id = :ticket_id;
	`
)

func (db *DB) deleteTicketTx(tx Tx, ticketID int) error {
	stmt, err := tx.PrepareNamed(deleteTicket)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"ticket_id": ticketID,
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

func (db *DB) GetOrderByID(orderID int) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"order_id": orderID,
		"iso8601":  ConstIso8061,
	}

	var order models.Order
	var user models.User
	var paymentBT []byte
	var ticketBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&order.ID,
		&order.Created,
		&order.Updated,
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&paymentBT,
		&ticketBT,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(paymentBT, &order.Payment); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ticketBT, &order.Tickets); err != nil {
		return nil, err
	}

	order.Client = &user

	return &order, nil
}

func (db *DB) GetOrderByExternalReference(externalReference string) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"external_reference": externalReference,
		"iso8601":            ConstIso8061,
	}

	var order models.Order
	var user models.User
	var paymentBT []byte
	var ticketBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&order.ID,
		&order.Created,
		&order.Updated,
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&paymentBT,
		&ticketBT,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(paymentBT, &order.Payment); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(ticketBT, &order.Tickets); err != nil {
		return nil, err
	}

	order.Client = &user

	return &order, nil
}
