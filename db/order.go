package db

import (
	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

type OrderStorage interface {
	InsertOrder(userID int, clientID int, events []models.Event, ticketsByEvent map[int]int) (*models.Order, error)
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
