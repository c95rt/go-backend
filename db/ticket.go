package db

import (
	"database/sql"
	"encoding/json"
	"strings"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

type TicketStorage interface {
	GetOrderByTicketUUID(ticketUUID string) (*models.Order, error)
	GetOrderByTicketID(id int) (*models.Order, error)
	UseTicket(ticketID int) error
	UpdateTicket(ticketID int, eventID int) error
	GetTickets(*models.GetTicketsOpts) (*models.GetTicketsStruct, error)
}

const (
	getOrderByTicketUUID = `
	SELECT
		ticket.id,
		ticket.uuid,
		ticket.used,
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
		ticket
	INNER JOIN
		orders ON (orders.id = ticket.order_id AND orders.active = true)
	INNER JOIN
		event ON (event.id = ticket.event_id AND event.active = true)
	WHERE
		ticket.uuid = :uuid AND
		ticket.active = true
	`

	getOrderByTicketID = `
	SELECT
		ticket.id,
		ticket.uuid,
		ticket.used,
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
		ticket
	INNER JOIN
		orders ON (orders.id = ticket.order_id AND orders.active = true)
	INNER JOIN
		event ON (event.id = ticket.event_id AND event.active = true)
	WHERE
		ticket.id = :id AND
		ticket.active = true
	`

	useTicket = `
	UPDATE
		ticket
	SET
		used = 1,
		updated = current_timestamp();
	WHERE
		ticket.id = :ticket_id AND
		ticket.active = true;
	`

	updateTicket = `
	UPDATE
		ticket
	SET
		event_id = :event_id,
		updated = current_timestamp()
	WHERE
		id = :ticket_id AND
		active = true AND
		used = 0;
	`

	getTickets = `
	SELECT
		ticket.id,
		ticket.uuid,
		ticket.used,
		IF(payment.id IS NOT NULL, true, false) AS paid,
		client.id,
		client.firstname,
		client.lastname,
		event.id,
		event.price,
		event.start_date_time,
		event.end_date_time
	FROM
		ticket
	INNER JOIN
		orders ON (orders.id = ticket.order_id)
	LEFT JOIN
		payment ON (payment.id = (SELECT id FROM payment WHERE order_id = orders.id AND status_id = :status_id ORDER BY id DESC LIMIT 1))
	INNER JOIN
		event ON (event.id = ticket.event_id)
	INNER JOIN
		user AS client ON (orders.client_id = client.id)
	WHERE
		ticket.active = true
		#FILTERS#
	ORDER BY
		event.start_date_time DESC
	LIMIT :limit_to OFFSET :limit_from
	`

	countTickets = `
	SELECT
		COUNT(ticket.id)
	FROM
		ticket
	INNER JOIN
		orders ON (orders.id = ticket.order_id)
	INNER JOIN
		event ON (event.id = ticket.event_id)
	INNER JOIN
		user AS client ON (orders.client_id = client.id)
	WHERE
		ticket.active = true
		#FILTERS#
	`
)

func (db *DB) GetOrderByTicketUUID(ticketUUID string) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByTicketUUID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"uuid":    ticketUUID,
		"iso8601": ConstIso8061,
	}

	var ticket models.Ticket
	var event models.Event
	var order models.Order
	var paymentBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&ticket.ID,
		&ticket.UUID,
		&ticket.Used,
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

func (db *DB) GetOrderByTicketID(id int) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByTicketID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"id":      id,
		"iso8601": ConstIso8061,
	}

	var ticket models.Ticket
	var event models.Event
	var order models.Order
	var paymentBT []byte

	row := stmt.QueryRow(args)
	if err := row.Scan(
		&ticket.ID,
		&ticket.UUID,
		&ticket.Used,
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

func (db *DB) UseTicket(ticketID int) error {
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

	err = db.useTicketTx(tx, ticketID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) useTicketTx(tx Tx, ticketID int) error {
	stmt, err := tx.PrepareNamed(useTicket)
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

func (db *DB) UpdateTicket(ticketID int, eventID int) error {
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

	err = db.updateTicketTx(tx, ticketID, eventID)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updateTicketTx(tx Tx, ticketID int, eventID int) error {
	stmt, err := tx.PrepareNamed(updateTicket)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"ticket_id": ticketID,
		"event_id":  eventID,
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

func (db *DB) GetTickets(opts *models.GetTicketsOpts) (*models.GetTicketsStruct, error) {
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
	if opts.TicketUUID != "" {
		filters += " AND ticket.uuid <= :ticket_uuid "
		args["ticket_uuid"] = opts.TicketUUID
	}
	if opts.LimitTo == 0 {
		opts.LimitTo = 10
	}
	args["limit_to"] = opts.LimitTo
	args["limit_from"] = opts.LimitFrom
	args["status_id"] = ConstPaymentStatuses.Approved.ID

	totalTickets, err := db.countTickets(filters, args)
	if err != nil {
		return nil, err
	}

	query := strings.ReplaceAll(getTickets, "#FILTERS#", filters)

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	tickets := models.GetTicketsStruct{
		Total: totalTickets,
	}

	for rows.Next() {
		var ticket models.Ticket
		var client models.User
		var event models.Event
		if err := rows.Scan(
			&ticket.ID,
			&ticket.UUID,
			&ticket.Used,
			&ticket.Paid,
			&client.ID,
			&client.Firstname,
			&client.Lastname,
			&event.ID,
			&event.Price,
			&event.StartDateTime,
			&event.EndDateTime,
		); err != nil {
			return nil, err
		}

		ticket.Client = &client
		ticket.Event = &event
		tickets.Tickets = append(tickets.Tickets, ticket)
	}

	return &tickets, nil
}

func (db *DB) countTickets(filters string, args map[string]interface{}) (int, error) {
	query := strings.ReplaceAll(countTickets, "#FILTERS#", filters)
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
