package db

import (
	"database/sql"
	"encoding/json"
	"strings"

	"bitbucket.org/parqueoasis/backend/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type UserStorage interface {
	InsertUser(*models.InsertAdminUserOpts) (int, error)
	UpdateUserPassword(*models.User) error
	GetUserByID(userID int) (*models.User, error)
	ValidateUserEmailAndDNI(email string, dni string) (emailCounter int, dniCounter int, err error)
	GetUsers(*models.GetUsersOpts) (*models.UsersStruct, error)
	UpdateUser(userID int, opts *models.UpdateUserOpts) error
}

const (
	insertUser = `
	INSERT
		user
	SET
		email = :email,
		password = :password,
		firstname = :firstname,
		lastname = :lastname
	`

	insertUserAdditional = `
	INSERT
		user_additional
	SET
		dni = :dni,
		phone = :phone,
		user_id = :user_id
	`

	updateUserPassword = `
	UPDATE
		user
	SET
		password = :password,
		remember_token = NULL
	WHERE
		user.id = :user_id AND
		user.active = 1
	`

	insertUserRoles = `
	INSERT INTO
		pivot_role_user (user_id, role_id)
	SELECT
		:user_id,
		role.id
	FROM
		role
	WHERE
		role.id IN (:role_ids)
	AND role.active = 1
	`

	getUserByID = `
	SELECT
		user.id,
		user.firstname,
		user.lastname,
		user.email,
		user.password,
		user.created,
		user.updated,
		user.active,
		user_additional.id,
		user_additional.phone,
		user_additional.dni,
		COALESCE(CONCAT('[',GROUP_CONCAT(JSON_OBJECT('id', role.id, 'name', role.name)),']'))
	FROM
		user
	INNER JOIN
		user_additional ON (user_additional.user_id = user.id)
	INNER JOIN
		pivot_role_user ON (pivot_role_user.user_id = user.id)
	INNER JOIN
		role ON (role.id = pivot_role_user.role_id AND role.active = 1)
	WHERE
		user.active = 1 AND
		user.id = :user_id
	GROUP BY
		user.id
	`

	getUsers = `
	SELECT
		user.id,
		user.firstname,
		user.lastname,
		user.email,
		user.password,
		user.created,
		user.updated,
		user.active,
		user_additional.id,
		user_additional.phone,
		user_additional.dni,
		COALESCE(CONCAT('[',GROUP_CONCAT(JSON_OBJECT('id', role.id, 'name', role.name)),']'))
	FROM
		user
	INNER JOIN
		pivot_role_user ON (pivot_role_user.user_id = user.id)
	INNER JOIN
		role ON (role.id = pivot_role_user.role_id AND role.active = 1)
	INNER JOIN
		user_additional ON (user_additional.user_id = user.id)
	WHERE
		user.active = 1
		#FILTERS#
	GROUP BY
		user.id
	ORDER BY
		user.id ASC
	LIMIT :limit_to OFFSET :limit_from
	`

	countUsers = `
	SELECT
		count(DISTINCT user.id)
	FROM
		user
	INNER JOIN pivot_role_user ON (pivot_role_user.user_id = user.id)
	INNER JOIN role ON (role.id = pivot_role_user.role_id AND role.active = 1)
	INNER JOIN user_additional ON (user_additional.user_id = user.id)
	WHERE
		user.active = 1
		#FILTERS#
	`
)

func (db *DB) InsertUser(opts *models.InsertAdminUserOpts) (int, error) {
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

	userID, err := db.insertUserTx(tx, opts)
	if err != nil {
		return 0, err
	}

	userAdditional := models.UserAdditional{
		ID:    userID,
		DNI:   opts.DNI,
		Phone: opts.Phone,
	}

	err = db.insertUserAdditionalTx(tx, userID, &userAdditional)
	if err != nil {
		return 0, err
	}

	err = db.insertUserRolesTx(tx, userID, opts.Roles)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func (db *DB) insertUserTx(tx Tx, opts *models.InsertAdminUserOpts) (int, error) {
	stmt, err := tx.PrepareNamed(insertUser)
	if err != nil {
		return 0, err
	}

	args := map[string]interface{}{
		"email":     opts.Email,
		"password":  opts.Password,
		"firstname": opts.Firstname,
		"lastname":  opts.Lastname,
	}

	result, err := stmt.Exec(args)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rowsAffected != 1 {
		return 0, errors.Errorf("expected %d and updated %d rows", 1, rowsAffected)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (db *DB) insertUserAdditionalTx(tx Tx, userID int, opts *models.UserAdditional) error {
	stmt, err := tx.PrepareNamed(insertUserAdditional)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"dni":     opts.DNI,
		"phone":   opts.Phone,
		"user_id": userID,
	}

	result, err := stmt.Exec(args)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.Errorf("expected %d and updated %d rows", 1, rowsAffected)
	}

	return nil
}

func (db *DB) insertUserRolesTx(tx Tx, userID int, roleIDs []int) error {
	args := map[string]interface{}{
		"user_id":  userID,
		"role_ids": roleIDs,
	}
	query, nargs, err := sqlx.Named(insertUserRoles, args)
	if err != nil {
		return err
	}

	query, nargs, err = sqlx.In(query, nargs...)
	if err != nil {
		return err
	}

	query = tx.Rebind(query)

	result, err := tx.Exec(query, nargs...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if int(rowsAffected) != len(roleIDs) {
		return errors.Errorf("expected %d and inserted %d", len(roleIDs), rowsAffected)
	}

	return nil
}

func (db *DB) UpdateUserPassword(user *models.User) error {
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

	err = db.updateUserPasswordTx(tx, user.ID, user.Password)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) updateUserPasswordTx(tx Tx, userID int, password string) error {
	stmt, err := tx.PrepareNamed(updateUserPassword)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"user_id":  userID,
		"password": password,
	}

	result, err := stmt.Exec(args)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		return errors.Errorf("expected %d and updated %d rows", 1, rowsAffected)
	}

	return nil
}

func (db *DB) GetUserByID(userID int) (*models.User, error) {
	stmt, err := db.PrepareNamed(getUserByID)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"user_id": userID,
	}

	row := stmt.QueryRow(args)

	var user models.User
	var additional models.UserAdditional
	var rolesBT []byte
	if err := row.Scan(
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&user.Password,
		&user.Created,
		&user.Updated,
		&user.Active,
		&additional.ID,
		&additional.Phone,
		&additional.DNI,
		&rolesBT,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	user.Additional = &additional
	if err := json.Unmarshal(rolesBT, &user.Roles); err != nil {
		return nil, err
	}

	return &user, nil
}

func (db *DB) GetUsers(opts *models.GetUsersOpts) (*models.UsersStruct, error) {
	var filters string
	args := make(map[string]interface{})
	if opts.CreatedFrom != "" {
		filters += " AND DATE(CONVERT_TZ(user.created, 'UTC', 'America/Santiago')) >= :created_from "
		args["created_from"] = opts.CreatedFrom
	}
	if opts.CreatedTo != "" {
		filters += " AND DATE(CONVERT_TZ(user.created, 'UTC', 'America/Santiago')) <= :created_to "
		args["created_to"] = opts.CreatedTo
	}
	if len(opts.UserIDs) != 0 {
		filters += " AND user.id IN (:user_ids) "
		args["user_ids"] = opts.UserIDs
	}
	if len(opts.RoleIDs) != 0 {
		filters += " AND pivot_role_user.role_id IN (:role_ids) "
		args["role_ids"] = opts.RoleIDs
	}
	if len(opts.Emails) != 0 {
		filters += " AND user.email IN (:emails) "
		args["emails"] = opts.Emails
	}
	if len(opts.Firstnames) != 0 {
		filters += " AND (user.firstname LIKE '%" + strings.Join(opts.Firstnames, "%' OR user.firstname LIKE '%") + "%')"
	}
	if len(opts.Lastnames) != 0 {
		filters += " AND (user.lastname LIKE '%" + strings.Join(opts.Lastnames, "%' OR user.lastname LIKE '%") + "%')"
	}
	if len(opts.Phones) != 0 {
		filters += " AND user_additional.phone IN (:phones) "
		args["phones"] = opts.Phones
	}
	if len(opts.DNIs) != 0 {
		filters += " AND user_additional.dni IN (:dnis) "
		args["dnis"] = opts.DNIs
	}
	if opts.LimitTo == 0 {
		opts.LimitTo = 10
	}
	args["limit_to"] = opts.LimitTo
	args["limit_from"] = opts.LimitFrom

	totalUsers, err := db.countUsers(filters, args)
	if err != nil {
		return nil, err
	}

	query := strings.ReplaceAll(getUsers, "#FILTERS#", filters)

	query, nargs, err := sqlx.Named(query, args)
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

	users := models.UsersStruct{
		Total: totalUsers,
	}
	for rows.Next() {
		var user models.User
		var additional models.UserAdditional
		var rolesBT []byte
		if err := rows.Scan(
			&user.ID,
			&user.Firstname,
			&user.Lastname,
			&user.Email,
			&user.Password,
			&user.Created,
			&user.Updated,
			&user.Active,
			&additional.ID,
			&additional.Phone,
			&additional.DNI,
			&rolesBT,
		); err != nil {
			return nil, err
		}

		user.Additional = &additional
		if err := json.Unmarshal(rolesBT, &user.Roles); err != nil {
			return nil, err
		}

		users.Users = append(users.Users, user)
	}
	return &users, nil
}

func (db *DB) countUsers(filters string, args map[string]interface{}) (int, error) {
	query := strings.ReplaceAll(countUsers, "#FILTERS#", filters)

	query, nargs, err := sqlx.Named(query, args)
	if err != nil {
		return 0, err
	}

	query, nargs, err = sqlx.In(query, nargs...)
	if err != nil {
		return 0, err
	}

	query = db.Rebind(query)

	row := db.QueryRow(query, nargs...)
	var total int
	if err := row.Scan(
		&total,
	); err != nil {
		return 0, err
	}

	return total, nil
}

func (db *DB) UpdateUser(userID int, opts *models.UpdateUserOpts) error {
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

	err = db.deleteUserRolesTx(tx, userID)
	if err != nil {
		return err
	}

	err = db.insertUserRolesTx(tx, userID, opts.Roles)
	if err != nil {
		return err
	}

	err = db.updateUserAdditionalTx(tx, userID, opts)
	if err != nil {
		return err
	}

	err = db.updateUserTx(tx, userID, opts)
	if err != nil {
		return err
	}

	return nil
}

const (
	deleteUserRoles = `
	DELETE FROM
		pivot_role_user
	WHERE
		pivot_role_user.user_id = :user_id;
	`
)

func (db *DB) deleteUserRolesTx(tx Tx, userID int) error {
	stmt, err := tx.PrepareNamed(deleteUserRoles)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"user_id": userID,
	}

	_, err = stmt.Exec(args)
	if err != nil {
		return err
	}

	return nil
}

const (
	updateUserAdditional = `
	UPDATE
		user_additional
	SET
		dni = :dni,
		phone = :phone,
		updated = current_timestamp()
	WHERE
		user_id = :user_id
	`
)

func (db *DB) updateUserAdditionalTx(tx Tx, userID int, opts *models.UpdateUserOpts) error {
	stmt, err := tx.PrepareNamed(updateUserAdditional)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"user_id": userID,
		"dni":     opts.DNI,
		"phone":   opts.Phone,
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

const (
	updateUser = `
	UPDATE
		user
	SET
		firstname = :firstname,
		lastname = :lastname,
		email = :email,
		password = :password,
		updated = current_timestamp()
	WHERE
		id = :user_id
	`
)

func (db *DB) updateUserTx(tx Tx, userID int, opts *models.UpdateUserOpts) error {
	stmt, err := tx.PrepareNamed(updateUser)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"user_id":   userID,
		"firstname": opts.Firstname,
		"lastname":  opts.Lastname,
		"email":     opts.Email,
		"password":  opts.Password,
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

const (
	validateUserEmailAndDNI = `
	SELECT
		(
			SELECT
				COUNT(user.id)
			FROM
				user
			WHERE
				user.email = :email
		) email_counter,
		(
			SELECT
				COUNT(user_additional.id)
			FROM
				user_additional
			WHERE
				user_additional.dni = :dni
		) dni_counter
	`
)

func (db *DB) ValidateUserEmailAndDNI(email string, dni string) (emailCounter int, dniCounter int, err error) {
	stmt, err := db.PrepareNamed(validateUserEmailAndDNI)
	if err != nil {
		return 0, 0, err
	}

	args := map[string]interface{}{
		"email": email,
		"dni":   dni,
	}

	row := stmt.QueryRow(args)

	if err := row.Scan(
		&emailCounter,
		&dniCounter,
	); err != nil {
		return 0, 0, err
	}

	return emailCounter, dniCounter, nil
}
