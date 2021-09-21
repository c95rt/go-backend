package db

import (
	"database/sql"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

type OrderStorage interface {
	InsertOrder(userID int, clientID int, events []models.Event, ticketsByEvent map[int]int) (*models.Order, error)
	GetOrderByOrderIDAndTicketID(orderID int, ticketID int) (*models.Order, error)
	DeleteTicket(ticketID int) error
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
		order_id = :order_id,
		event_id = :event_id
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
			ticket.ID, err = db.insertTicketTx(tx, orderID, event.ID)
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

func (db *DB) insertTicketTx(tx Tx, orderID int, eventID int) (int, error) {
	stmt, err := tx.PrepareNamed(insertTicket)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
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

const (
	getOrderByOrderIDAndTicketID = `
	SELECT
		ticket.id,
		event.id,
		event.start_date_time,
		event.end_date_time,
		orders.id,
		orders.paid,
		orders.created,
		orders.updated
	FROM
		ticket
	INNER JOIN
		orders ON (orders.id = ticket.order_id AND orders.paid = true AND orders.active = true)
		event ON (event.id = ticket.event_id AND event.active = true)
	WHERE
		ticket.id = :ticket_id AND
		ticket.order_id = :order_id AND
		ticket.active = true
	`
)

func (db *DB) GetOrderByOrderIDAndTicketID(orderID int, ticketID int) (*models.Order, error) {
	stmt, err := db.PrepareNamed(getOrderByOrderIDAndTicketID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"ticket_id": ticketID,
		"order_id":  orderID,
	}

	var ticket models.Ticket
	var event models.Event
	var order models.Order
	row := stmt.QueryRow(args)
	if err := row.Scan(
		&ticket.ID,
		&event.ID,
		&event.StartDateTime,
		&event.EndDateTime,
		&order.ID,
		&order.Paid,
		&order.Created,
		&order.Updated,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
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
	DELETE FROM
		ticket
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
