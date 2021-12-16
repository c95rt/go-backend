package db

import (
	"database/sql"
	"fmt"
	"strings"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type EventStorage interface {
	InsertEvents(*models.InsertEventsOpts) error
	GetEventByID(eventID int) (*models.Event, error)
	GetEventsByIDs(eventIDs []int) ([]models.Event, error)
	GetEvents(*models.GetEventsOpts) (*models.EventsStruct, error)
	GetEventTypes() ([]models.EventType, error)
}

const (
	insertEvents = `
	INSERT INTO
		event (name, event_type_id, start_date_time, end_date_time, price)
	VALUES
		%s
	`

	getEventByID = `
	SELECT
		event.id,
		event.name,
		event_type.id,
		event_type.name,
		event.start_date_time,
		event.end_date_time,
		event.price,
		event.created,
		event.updated
	FROM
		event
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	WHERE
		event.active = 1 AND
		event.id = :event_id
	`

	getEventsByIDs = `
	SELECT
		event.id,
		event.name,
		event_type.id,
		event_type.name,
		event.start_date_time,
		event.end_date_time,
		event.price,
		event.created,
		event.updated
	FROM
		event
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	WHERE
		event.active = 1 AND
		event.id IN (:event_ids)
	`

	getEvents = `
	SELECT
		event.id,
		event.name,
		event_type.id,
		event_type.name,
		event.start_date_time,
		event.end_date_time,
		event.price,
		event.created,
		event.updated
	FROM
		event
	INNER JOIN
		event_type ON (event_type.id = event.event_type_id)
	WHERE
		event.active = 1
		#FILTERS#
	ORDER BY
		event.start_date_time ASC
	LIMIT :limit_to OFFSET :limit_from
	`

	countEvents = `
	SELECT
		COUNT(id)
	FROM
		event
	WHERE
		event.active = 1
		#FILTERS#
	`
)

func (db *DB) InsertEvents(opts *models.InsertEventsOpts) error {
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

	err = db.insertEventsTx(tx, opts)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) insertEventsTx(tx Tx, opts *models.InsertEventsOpts) error {
	var paramsArr []string
	var argsArr []interface{}

	for _, eventDate := range opts.Dates {
		for _, eventDateTime := range eventDate.Times {
			paramsArr = append(paramsArr, "(?, ?,?,?,?)")
			argsArr = append(argsArr, opts.Name, opts.TypeID, fmt.Sprintf("%s %s", eventDate.Date, eventDateTime.StartTime), fmt.Sprintf("%s %s", eventDate.Date, eventDateTime.EndTime), eventDateTime.Price)
		}
	}

	query := fmt.Sprintf(insertEvents, strings.Join(paramsArr, ","))
	result, err := tx.Exec(query, argsArr...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if int(rowsAffected) != len(paramsArr) {
		return errors.Errorf("expected %d and inserted %d", len(paramsArr), rowsAffected)
	}

	return nil
}

func (db *DB) GetEventByID(eventID int) (*models.Event, error) {
	stmt, err := db.PrepareNamed(getEventByID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"event_id": eventID,
	}

	row := stmt.QueryRow(args)

	var event models.Event
	var eventType models.EventType
	if err := row.Scan(
		&event.ID,
		&event.Name,
		&eventType.ID,
		&eventType.Name,
		&event.StartDateTime,
		&event.EndDateTime,
		&event.Price,
		&event.Created,
		&event.Updated,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	event.Type = &eventType

	return &event, nil
}

func (db *DB) GetEventsByIDs(eventIDs []int) ([]models.Event, error) {
	args := map[string]interface{}{
		"event_ids": eventIDs,
	}
	query, nargs, err := sqlx.Named(getEventsByIDs, args)
	if err != nil {
		return nil, err
	}

	query, nargs, err = sqlx.In(query, nargs...)
	if err != nil {
		return nil, err
	}

	query = db.Rebind(query)

	rows, err := db.Query(query, nargs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var event models.Event
		var eventType models.EventType
		if err := rows.Scan(
			&event.ID,
			&event.Name,
			&eventType.ID,
			&eventType.Name,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Price,
			&event.Created,
			&event.Updated,
		); err != nil {
			return nil, err
		}

		event.Type = &eventType
		events = append(events, event)
	}

	return events, nil
}

func (db *DB) GetEvents(opts *models.GetEventsOpts) (*models.EventsStruct, error) {
	var filters string
	args := make(map[string]interface{})
	if opts.Date == "" {
		filters += " AND event.start_date_time >= CONVERT_TZ(current_timestamp(), 'UTC', 'America/Santiago')"
	} else {
		filters += " AND DATE(event.start_date_time) = :date"
		args["date"] = opts.Date
	}
	if opts.TypeID != 0 {
		filters += " AND event.event_type_id = :type_id"
		args["type_id"] = opts.TypeID
	}
	if opts.LimitTo == 0 {
		opts.LimitTo = 10
	}
	args["limit_to"] = opts.LimitTo
	args["limit_from"] = opts.LimitFrom

	totalEvents, err := db.countEvents(filters, args)
	if err != nil {
		return nil, err
	}

	query := strings.ReplaceAll(getEvents, "#FILTERS#", filters)
	stmt, err := db.PrepareNamed(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query(args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := models.EventsStruct{
		Total: totalEvents,
	}
	for rows.Next() {
		var event models.Event
		var eventType models.EventType
		if err := rows.Scan(
			&event.ID,
			&event.Name,
			&eventType.ID,
			&eventType.Name,
			&event.StartDateTime,
			&event.EndDateTime,
			&event.Price,
			&event.Created,
			&event.Updated,
		); err != nil {
			return nil, err
		}

		event.Type = &eventType
		events.Events = append(events.Events, event)
	}

	return &events, nil
}

func (db *DB) countEvents(filters string, args map[string]interface{}) (int, error) {
	query := strings.ReplaceAll(countEvents, "#FILTERS#", filters)
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

const (
	getEventTypes = `
	SELECT
		event_type.id,
		event_type.name
	FROM
		event_type
	`
)

func (db *DB) GetEventTypes() ([]models.EventType, error) {
	rows, err := db.Query(getEventTypes)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var eventTypes []models.EventType
	for rows.Next() {
		var eventType models.EventType
		if err := rows.Scan(
			&eventType.ID,
			&eventType.Name,
		); err != nil {
			return nil, err
		}

		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes, nil
}
