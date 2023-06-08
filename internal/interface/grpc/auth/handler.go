// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"context"

	app "github.com/isaqueveras/powersso/internal/application/auth"
	domain "github.com/isaqueveras/powersso/internal/domain/auth"
	"github.com/isaqueveras/powersso/internal/utils"
	"github.com/isaqueveras/powersso/pkg/oops"
)

// Server implements proto interface
type Server struct {
	UnimplementedAuthenticationServer
}

// RegisterUser register user
func (s *Server) RegisterUser(ctx context.Context, in *User) (_ *Empty, err error) {
	if err = app.CreateAccount(ctx, &domain.CreateAccount{
		FirstName: utils.Pointer(in.FirstName),
		LastName:  utils.Pointer(in.LastName),
		Email:     utils.Pointer(in.Email),
		Password:  utils.Pointer(in.Password),
	}); err != nil {
		return nil, oops.HandlingGRPC(err)
	}

	return
}
