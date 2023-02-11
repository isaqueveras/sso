// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"

	"github.com/isaqueveras/power-sso/config"
	domain "github.com/isaqueveras/power-sso/internal/domain/auth"
	"github.com/isaqueveras/power-sso/internal/domain/auth/roles"
	domainUser "github.com/isaqueveras/power-sso/internal/domain/auth/user"
	"github.com/isaqueveras/power-sso/internal/infrastructure/auth"
	infraRoles "github.com/isaqueveras/power-sso/internal/infrastructure/auth/roles"
	infraSession "github.com/isaqueveras/power-sso/internal/infrastructure/auth/session"
	infraUser "github.com/isaqueveras/power-sso/internal/infrastructure/auth/user"
	"github.com/isaqueveras/power-sso/pkg/conversor"
	"github.com/isaqueveras/power-sso/pkg/database/postgres"
	"github.com/isaqueveras/power-sso/pkg/mailer"
	"github.com/isaqueveras/power-sso/pkg/oops"
	"github.com/isaqueveras/power-sso/tokens"
)

// Register is the business logic for the user register
func Register(ctx context.Context, in *RegisterRequest) error {
	transaction, err := postgres.NewTransaction(ctx, false)
	if err != nil {
		return oops.Err(err)
	}
	defer transaction.Rollback()

	if err = in.Prepare(); err != nil {
		return oops.Err(err)
	}

	if in.Roles == nil {
		in.Roles = new(roles.Roles)
		in.Roles.Add(roles.ReadActivationToken)
	}
	in.Roles.Parse()

	var (
		exists      bool
		userID      *string
		data        *domain.Register
		accessToken = new(string)
		repo        = auth.New(transaction, mailer.Client(config.Get()))
		repoUser    = infraUser.New(transaction)
	)

	if exists, err = repoUser.FindByEmailUserExists(in.Email); err != nil {
		return oops.Err(err)
	}

	if exists {
		return oops.Err(domain.ErrUserExists())
	}

	if data, err = conversor.TypeConverter[domain.Register](&in); err != nil {
		return oops.Err(err)
	}

	data.Roles = &in.Roles.String
	if userID, err = repo.Register(data); err != nil {
		return oops.Err(err)
	}

	if accessToken, err = repo.CreateAccessToken(userID); err != nil {
		return oops.Err(err)
	}

	if err = repo.SendMailActivationAccount(in.Email, accessToken); err != nil {
		return oops.Err(err)
	}

	in.SanitizePassword()
	if err = transaction.Commit(); err != nil {
		return oops.Err(err)
	}

	return nil
}

// Activation is the business logic for the user activation
func Activation(ctx context.Context, token *string) (err error) {
	transaction, err := postgres.NewTransaction(ctx, false)
	if err != nil {
		return oops.Err(err)
	}
	defer transaction.Rollback()

	var (
		repo        = auth.New(transaction, nil)
		repoUser    = infraUser.New(transaction)
		activeToken *domain.ActivateAccountToken
	)

	if activeToken, err = repo.GetActivateAccountToken(token); err != nil {
		return oops.Err(err)
	}

	if *activeToken.Used || !*activeToken.IsValid {
		return oops.Err(domain.ErrTokenIsNotValid())
	}

	user := domainUser.User{
		ID: activeToken.UserID,
	}

	if err = repoUser.GetUser(&user); err != nil {
		return oops.Err(err)
	}

	if !roles.Exists(roles.ReadActivationToken, roles.Roles{String: *user.Roles}) {
		return oops.Err(domain.ErrNotHavePermissionActiveAccount())
	}

	repoRoles := infraRoles.New(transaction)
	if err = repoRoles.RemoveRoles(user.ID, roles.ReadActivationToken); err != nil {
		return oops.Err(err)
	}

	rolesSession := roles.MakeEmptyRoles()
	rolesSession.Add(
		roles.ReadSession,
		roles.CreateSession,
		roles.DeleteSession,
	)
	rolesSession.ParseString()

	if err = repoRoles.AddRoles(user.ID, rolesSession.Strings()); err != nil {
		return oops.Err(err)
	}

	if err = repo.MarkTokenAsUsed(activeToken.ID); err != nil {
		return oops.Err(err)
	}

	if err = transaction.Commit(); err != nil {
		return oops.Err(err)
	}

	return nil
}

// Login is the business logic for the user login
func Login(ctx context.Context, in *LoginRequest) (*SessionResponse, error) {
	var (
		transaction *postgres.DBTransaction
		err         error
	)

	if transaction, err = postgres.NewTransaction(ctx, false); err != nil {
		return nil, oops.Err(err)
	}
	defer transaction.Rollback()

	var (
		cfg              = config.Get()
		repo             = auth.New(transaction, nil)
		repoUser         = infraUser.New(transaction)
		repoSession      = infraSession.New(transaction)
		user             = &domainUser.User{Email: in.Email}
		token            string
		passw, sessionID *string
	)

	if err = repoUser.GetUser(user); err != nil {
		return nil, oops.Err(err)
	}

	if user.IsActive != nil && !*user.IsActive {
		return nil, oops.Err(domain.ErrUserNotExists())
	}

	if user.BlockedTemporarily != nil && *user.BlockedTemporarily {
		return nil, oops.Err(domain.ErrUserBlockedTemporarily())
	}

	if passw, err = repo.Login(in.Email); err != nil {
		return nil, oops.Err(domain.ErrUserNotExists())
	}

	if err = in.ComparePasswords(passw, user.TokenKey); err != nil {
		var errAttempts error
		if errAttempts = repo.AddNumberFailedAttempts(user.ID); errAttempts != nil {
			return nil, oops.Err(errAttempts)
		}
		if errAttempts = transaction.Commit(); errAttempts != nil {
			return nil, oops.Err(errAttempts)
		}
		return nil, oops.Err(err)
	}

	if !roles.Exists(roles.CreateSession, roles.Roles{String: *user.Roles}) {
		return nil, oops.Err(domain.ErrNotHavePermissionLogin())
	}

	if sessionID, err = repoSession.Create(user.ID, &in.ClientIP, &in.UserAgent); err != nil {
		return nil, oops.Err(err)
	}

	if token, err = tokens.NewUserAuthToken(cfg, user, sessionID); err != nil {
		return nil, oops.Err(err)
	}

	userRoles := roles.MakeEmptyRoles()
	userRoles.String = *user.Roles
	userRoles.ParseArray()

	res := &SessionResponse{
		SessionID:   sessionID,
		Level:       user.UserType,
		UserID:      user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Roles:       userRoles.Arrays(),
		About:       user.About,
		AvatarURL:   user.Avatar,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt,
		Token:       &token,
		// TODO: implement the extra user data in the database
		RawData: make(map[string]any),
	}

	if err = transaction.Commit(); err != nil {
		return nil, oops.Err(err)
	}

	return res, nil
}

// Logout is the business logic for the user logout
func Logout(ctx context.Context, sessionID *string) (err error) {
	var transaction *postgres.DBTransaction
	if transaction, err = postgres.NewTransaction(ctx, false); err != nil {
		return oops.Err(err)
	}
	defer transaction.Rollback()

	if err = infraSession.
		New(transaction).
		Delete(sessionID); err != nil {
		return oops.Err(err)
	}

	if err = transaction.Commit(); err != nil {
		return oops.Err(err)
	}

	return
}
