// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	"github.com/isaqueveras/power-sso/pkg/i18n"
	"github.com/isaqueveras/power-sso/pkg/oops"
)

// ErrUserExists creates and returns an error when the user already exists
func ErrUserExists() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_user_exists"), http.StatusBadRequest)
}

// ErrNotHavePermissionActiveAccount creates and returns an error when the user does not have permission to active the account
func ErrNotHavePermissionActiveAccount() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_not_have_permission_active_account"), http.StatusBadRequest)
}

// ErrTokenIsNotValid creates and returns an error when the token is not valid
func ErrTokenIsNotValid() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_token_is_not_valid"), http.StatusBadRequest)
}

// ErrNotHavePermissionLogin creates and returns an error when the user does not have permission to login
func ErrNotHavePermissionLogin() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_not_have_permission_login"), http.StatusBadRequest)
}

// ErrUserNotExists creates and returns an error when the user does not exists
func ErrUserNotExists() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_user_not_exists"), http.StatusNotFound)
}

// ErrEmailOrPasswordIsNotValid creates and returns an error when the email or password is not valid
func ErrEmailOrPasswordIsNotValid() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_email_or_password_is_not_valid"), http.StatusBadRequest)
}

// ErrUserBlockedTemporarily creates and returns an error when the user is blocked temporarily
func ErrUserBlockedTemporarily() *oops.Error {
	return oops.NewError(i18n.Value("errors.handling.err_user_blocked_temporarily"), http.StatusForbidden)
}