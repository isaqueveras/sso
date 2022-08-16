// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package user

import (
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/isaqueveras/power-sso/internal/domain/user"
	"github.com/isaqueveras/power-sso/pkg/database/postgres"
	"github.com/isaqueveras/power-sso/pkg/oops"
)

// pgUser is the implementation
// of transaction for the user repository
type pgUser struct {
	DB *postgres.DBTransaction
}

// findByEmailUserExists check if the user exists by email in the database
func (pg *pgUser) findByEmailUserExists(email *string) (exists bool, err error) {
	if err = pg.DB.Builder.
		Select("COUNT(id) > 0").
		From("users").
		Where(squirrel.Eq{
			"email": email,
		}).
		Scan(&exists); err != nil && err != sql.ErrNoRows {
		return false, oops.Err(err)
	}

	return
}

// getUser get the user from the database
func (pg *pgUser) getUser(data *user.User) (err error) {
	var where squirrel.Eq
	if data.ID != nil {
		where = squirrel.Eq{"id": data.ID}
	} else if data.Email != nil {
		where = squirrel.Eq{"email": data.Email}
	}

	if err = pg.DB.Builder.
		Select(`
			id,
			email,
			first_name,
			last_name,
			roles,
			about,
			avatar,
			phone_number,
			address,
			city,
			country,
			gender,
			postcode,
			token_key,
			birthday,
			created_at,
			updated_at,
			login_date`).
		From("users").
		Where(where).
		Scan(&data.ID, &data.Email, &data.FirstName, &data.LastName, &data.Roles,
			&data.About, &data.Avatar, &data.PhoneNumber, &data.Address, &data.City,
			&data.Country, &data.Gender, &data.Postcode, &data.TokenKey, &data.Birthday,
			&data.CreatedAt, &data.UpdatedAt, &data.LoginDate); err != nil {
		return oops.Err(err)
	}
	return nil
}
