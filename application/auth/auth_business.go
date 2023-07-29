// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/isaqueveras/powersso/database/postgres"
	domain "github.com/isaqueveras/powersso/domain/auth"
	infra "github.com/isaqueveras/powersso/infrastructure/persistencie/auth"
	"github.com/isaqueveras/powersso/mailer"
	"github.com/isaqueveras/powersso/oops"
	"github.com/isaqueveras/powersso/tokens"
	"github.com/isaqueveras/powersso/utils"
)

// CreateAccount is the business logic for the user register
func CreateAccount(ctx context.Context, in *domain.CreateAccount) error {
	tx, err := postgres.NewTransaction(ctx, false)
	if err != nil {
		return oops.Err(err)
	}
	defer tx.Rollback()

	if err = in.Prepare(); err != nil {
		return oops.Err(err)
	}

	repoAuth := infra.NewAuthRepository(tx, mailer.Client())
	repoUser := infra.NewUserRepository(tx)

	if err = repoUser.Exist(in.Email); err != nil {
		return oops.Err(err)
	}

	var userID *uuid.UUID
	if userID, err = repoAuth.CreateAccount(in); err != nil {
		return oops.Err(err)
	}

	var token *uuid.UUID
	if token, err = repoAuth.CreateAccessToken(userID); err != nil {
		return oops.Err(err)
	}

	if err = repoAuth.SendMailActivationAccount(in.Email, token); err != nil {
		return oops.Err(err)
	}

	if err = tx.Commit(); err != nil {
		return oops.Err(err)
	}

	return nil
}

// Activation is the business logic for the user activation
func Activation(ctx context.Context, token *uuid.UUID) (err error) {
	var tx *postgres.Transaction
	if tx, err = postgres.NewTransaction(ctx, false); err != nil {
		return oops.Err(err)
	}
	defer tx.Rollback()

	var (
		repoAuth = infra.NewAuthRepository(tx, nil)
		repoUser = infra.NewUserRepository(tx)
		repoRole = infra.NewFlagRepo(tx)
	)

	activeToken := &domain.ActivateAccount{ID: token}
	if err = repoAuth.GetActivateAccountToken(activeToken); err != nil {
		return oops.Err(err)
	}

	if !activeToken.IsValid() {
		return oops.Err(domain.ErrTokenIsNotValid())
	}

	user := domain.User{ID: activeToken.UserID}
	if err = repoUser.Get(&user); err != nil {
		return oops.Err(err)
	}

	if err = repoRole.Set(user.ID, utils.Pointer(domain.FlagEnabledAccount)); err != nil {
		return oops.Err(err)
	}

	if err = repoAuth.MarkTokenAsUsed(activeToken.ID); err != nil {
		return oops.Err(err)
	}

	if err = tx.Commit(); err != nil {
		return oops.Err(err)
	}

	return nil
}

// Login is the business logic for the user login
func Login(ctx context.Context, in *domain.Login) (*domain.Session, error) {
	tx, err := postgres.NewTransaction(ctx, false)
	if err != nil {
		return nil, oops.Err(err)
	}
	defer tx.Rollback()

	var (
		repoAuth    = infra.NewAuthRepository(tx, nil)
		repoUser    = infra.NewUserRepository(tx)
		repoSession = infra.NewSessionRepository(tx)
	)

	user := &domain.User{Email: in.Email}
	if err = repoUser.Get(user); err != nil {
		return nil, oops.Err(err)
	}

	if !user.HasFlag(domain.FlagEnabledAccount) {
		return nil, oops.Err(domain.ErrNotHavePermissionLogin())
	}

	if !user.IsActive() {
		return nil, oops.Err(domain.ErrUserNotExists())
	}

	if user.IsBlocked() {
		return nil, oops.Err(domain.ErrUserBlockedTemporarily())
	}

	if err = in.ComparePasswords(user.Password, user.Key); err != nil {
		if errAttempts := repoAuth.AddAttempts(user.ID); errAttempts != nil {
			return nil, oops.Err(errAttempts)
		}
		if errAttempts := tx.Commit(); errAttempts != nil {
			return nil, oops.Err(errAttempts)
		}
		return nil, oops.Err(err)
	}

	if user.OTPConfigured() {
		if err = utils.ValidateToken(user.OTPToken, in.OTP); err != nil {
			return nil, oops.Err(domain.ErrOTPTokenInvalid())
		}
	}

	var sessionID *uuid.UUID
	if sessionID, err = repoSession.Create(user.ID, in.ClientIP, in.UserAgent); err != nil {
		return nil, oops.Err(err)
	}

	var token *string
	if token, err = tokens.NewUserAuthToken(user, sessionID); err != nil {
		return nil, oops.Err(err)
	}

	if err = tx.Commit(); err != nil {
		return nil, oops.Err(err)
	}

	return &domain.Session{
		SessionID: sessionID,
		Level:     user.Level,
		UserID:    user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		Token:     token,
		RawData:   make(map[string]any),
	}, nil
}

// Logout is the business logic for the user logout
func Logout(ctx context.Context, sessionID *uuid.UUID) (err error) {
	var tx *postgres.Transaction
	if tx, err = postgres.NewTransaction(ctx, false); err != nil {
		return oops.Err(err)
	}
	defer tx.Rollback()

	if err = infra.NewSessionRepository(tx).Delete(sessionID); err != nil {
		return oops.Err(err)
	}

	if err = tx.Commit(); err != nil {
		return oops.Err(err)
	}

	return
}

// LoginSteps is the business logic needed to retrieve needed steps for log a user in
func LoginSteps(ctx context.Context, email *string) (res *domain.Steps, err error) {
	var tx *postgres.Transaction
	if tx, err = postgres.NewTransaction(ctx, true); err != nil {
		return nil, oops.Err(err)
	}
	defer tx.Rollback()

	repository := infra.NewAuthRepository(tx, nil)
	return repository.LoginSteps(email)
}
