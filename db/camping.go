package db

import (
	"strings"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/pkg/errors"
)

type CampingStorage interface {
	InsertCamping(clientID int, eventID int, tickets int, price int) (*models.Camping, error)
	GetCampings(opts *models.GetCampingsOpts) (*models.GetCampingsStruct, error)
}

const (
	insertCamping = `
	INSERT
		camping
	SET
		event_id = :event_id,
		client_id = :client_id,
		tickets = :tickets,
		transaction_id = :transaction_id,
		price = :price
	`

	getCampings = `
	SELECT
		camping.id,
		camping.transaction_id,
		camping.tickets,
		camping.price,
		camping.created,
		camping.updated,
		client.id,
		client.firstname,
		client.lastname,
		client.email,
		event.id,
		event.name,
		event.start_date_time,
		event.end_date_time,
		event.price,
		event_type.id,
		event_type.name
	FROM
		camping
	INNER JOIN
		event ON (event.id = camping.event_id AND event.event_type_id = 2)
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	INNER JOIN
		user AS client ON (camping.client_id = client.id)
	WHERE
		camping.active = true
		#FILTERS#
	ORDER BY
		event.start_date_time DESC
	LIMIT :limit_to OFFSET :limit_from
	`

	countCampings = `
	SELECT
		COUNT(camping.id)
	FROM
		camping
	INNER JOIN
		event ON (event.id = camping.event_id AND event.event_type_id = 2)
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	INNER JOIN
		user AS client ON (camping.client_id = client.id)
	WHERE
		camping.active = true
		#FILTERS#
	`
)

func (db *DB) InsertCamping(clientID int, eventID int, tickets int, price int) (*models.Camping, error) {
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

	transactionID := GenerateCampingUUID()

	campingID, newErr := db.insertCampingTx(tx, clientID, eventID, transactionID, tickets, price)
	if newErr != nil {
		err = newErr
		return nil, err
	}

	camping := models.Camping{
		ID: campingID,
		Client: &models.User{
			ID: clientID,
		},
		Tickets:       tickets,
		TransactionID: transactionID,
	}

	return &camping, nil
}

func (db *DB) insertCampingTx(tx Tx, clientID int, eventID int, transactionID string, tickets int, price int) (int, error) {
	stmt, err := tx.PrepareNamed(insertCamping)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
		"event_id":       eventID,
		"client_id":      clientID,
		"tickets":        tickets,
		"transaction_id": transactionID,
		"price":          price * tickets,
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

func (db *DB) GetCampings(opts *models.GetCampingsOpts) (*models.GetCampingsStruct, error) {
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
		filters += " AND camping.transaction_id = :transaction_id "
		args["transaction_id"] = opts.TransactionID
	}
	if opts.ClientID != 0 {
		filters += " AND camping.client_id = :client_id "
		args["client_id"] = opts.ClientID
	}
	if opts.LimitTo == 0 {
		opts.LimitTo = 10
	}
	args["limit_to"] = opts.LimitTo
	args["limit_from"] = opts.LimitFrom

	totalCampings, err := db.countCampings(filters, args)
	if err != nil {
		return nil, err
	}

	query := strings.ReplaceAll(getCampings, "#FILTERS#", filters)

	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	campings := models.GetCampingsStruct{
		Total: totalCampings,
	}

	for rows.Next() {
		var camping models.Camping
		var client models.User
		var event models.Event
		var eventType models.EventType
		if err := rows.Scan(
			&camping.ID,
			&camping.TransactionID,
			&camping.Tickets,
			&camping.Price,
			&camping.Created,
			&camping.Updated,
			&client.ID,
			&client.Firstname,
			&client.Lastname,
			&client.Email,
			&event.ID,
			&event.Name,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Price,
			&eventType.ID,
			&eventType.Name,
		); err != nil {
			return nil, err
		}

		camping.Client = &client
		camping.Event = &event

		campings.Campings = append(campings.Campings, camping)
	}

	return &campings, nil
}

func (db *DB) countCampings(filters string, args map[string]interface{}) (int, error) {
	query := strings.ReplaceAll(countCampings, "#FILTERS#", filters)
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
