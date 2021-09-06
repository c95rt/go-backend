package db

import (
	"database/sql"
	"encoding/json"

	"bitbucket.org/parqueoasis/backend/models"
)

type AuthStorage interface {
	GetUserLoginByEmail(string) (*models.User, error)
	GetUserByRememberToken(string) (*models.User, error)
	UpdateUserRememberToken(int, string) error
}

const (
	getUserLoginByEmail = `
	SELECT
		user.id,
		user.firstname,
		user.lastname,
		user.email,
		user.password,
		user.created,
		user.updated,
		user.active,
		COALESCE(CONCAT('[',GROUP_CONCAT(JSON_OBJECT('id', role.id, 'name', role.name)),']'), '[]')
	FROM user
	INNER JOIN pivot_role_user ON (pivot_role_user.user_id = user.id)
	INNER JOIN role ON (role.id = pivot_role_user.role_id AND role.active = 1)
	WHERE user.email IN (:email)
	AND user.active = 1
	GROUP BY user.id
	`

	getUserByRememberToken = `
	SELECT
		user.id,
		user.email
	FROM user
	WHERE user.active = 1
	AND user.remember_token = :remember_token
	`

	updateUserRememberToken = `
	UPDATE
		user
	SET
		remember_token = :token
	WHERE
		id = :user_id
	`
)

func (db *DB) GetUserLoginByEmail(email string) (*models.User, error) {
	stmt, err := db.PrepareNamed(getUserLoginByEmail)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"email": email,
	}

	row := stmt.QueryRow(args)

	var user models.User
	var rolesBytes []byte

	if err := row.Scan(
		&user.ID,
		&user.Firstname,
		&user.Lastname,
		&user.Email,
		&user.Password,
		&user.Created,
		&user.Updated,
		&user.Active,
		&rolesBytes,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	var roles []models.Role
	err = json.Unmarshal(rolesBytes, &roles)
	if err != nil {
		return nil, err
	}
	user.Roles = roles

	return &user, nil
}

func (db *DB) GetUserByRememberToken(token string) (*models.User, error) {
	stmt, err := db.PrepareNamed(getUserByRememberToken)
	if err != nil {
		return nil, err
	}

	args := map[string]interface{}{
		"remember_token": token,
	}

	row := stmt.QueryRow(args)

	var user models.User

	if err := row.Scan(
		&user.ID,
		&user.Email,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (db *DB) UpdateUserRememberToken(userID int, token string) error {
	stmt, err := db.PrepareNamed(updateUserRememberToken)
	if err != nil {
		return err
	}

	args := map[string]interface{}{
		"token":   token,
		"user_id": userID,
	}

	_, err = stmt.Exec(args)
	if err != nil {
		return err
	}

	return nil
}
