// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package postgres

import (
	"database/sql"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"

	pg "github.com/isaqueveras/powersso/database/postgres"
	"github.com/isaqueveras/powersso/domain/auth"
	"github.com/isaqueveras/powersso/oops"
	"github.com/isaqueveras/powersso/utils"
)

// PGAuth is the implementation of transaction for the auth repository
type PGAuth struct {
	DB *pg.Transaction
}

// CreateAccount register the user in the database
func (pg *PGAuth) CreateAccount(input *auth.CreateAccount) (userID *uuid.UUID, err error) {
	_cols, _vals, err := utils.FormatValuesInUp(input)
	if err != nil {
		return nil, oops.Err(err)
	}

	if err = pg.DB.Builder.
		Insert("users").
		Columns(_cols...).
		Values(_vals...).
		Suffix(`RETURNING "id"`).
		Scan(&userID); err != nil {
		return nil, oops.Err(err)
	}

	return
}

// CreateAccessToken create the access token for the user
func (pg *PGAuth) CreateAccessToken(userID *uuid.UUID) (token *uuid.UUID, err error) {
	if err = pg.DB.Builder.
		Insert("activate_account_tokens").
		Columns("user_id", "expires_at").
		Values(userID, time.Now().Add(30*time.Minute)).
		Suffix(`RETURNING "id"`).
		Scan(&token); err != nil {
		return token, oops.Err(err)
	}

	return
}

// GetActivateAccountToken get the activate account token from the database
func (pg *PGAuth) GetActivateAccountToken(data *auth.ActivateAccount) (err error) {
	if err = pg.DB.Builder.
		Select("user_id, used, expires_at >= now(), expires_at, created_at").
		From("activate_account_tokens").
		Where("id = ?", data.ID).
		Limit(1).
		Scan(&data.UserID, &data.Used, &data.Valid, &data.ExpiresAt, &data.CreatedAt); err != nil && err != sql.ErrNoRows {
		return oops.Err(err)
	}

	return
}

// MarkTokenAsUsed mark the token as used in the database
func (pg *PGAuth) MarkTokenAsUsed(token *uuid.UUID) (err error) {
	if _, err = pg.DB.Builder.
		Update("activate_account_tokens").
		Set("used", true).
		Where("id = ?", token).
		Exec(); err != nil {
		return oops.Err(err)
	}

	return
}

func (pg *PGAuth) AddAttempts(userID *uuid.UUID) (err error) {
	if _, err = pg.DB.Builder.
		Update("users").
		Set("attempts", squirrel.Expr("attempts + 1")).
		Set("last_failure", squirrel.Expr("NOW()")).
		Where("id = ?", userID).
		Exec(); err != nil && err != sql.ErrNoRows {
		return oops.Err(err)
	}

	return
}

func (pg *PGAuth) LoginSteps(email *string) (steps *auth.Steps, err error) {
	steps = new(auth.Steps)
	if err = pg.DB.Builder.
		Select("first_name").
		Column("(flag&?) <> 0 AND 	(flag&?) <> 0",
			auth.FlagOTPEnable, auth.FlagOTPSetup).
		From("users").
		Where("email = ?", email).
		Limit(1).
		Scan(&steps.Name, &steps.OTP); err != nil && err != sql.ErrNoRows {
		return nil, oops.Err(err)
	}

	return
}
