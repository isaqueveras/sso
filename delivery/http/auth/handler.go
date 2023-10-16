// Copyright (c) 2022 Isaque Veras
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gopowersso "github.com/isaqueveras/go-powersso"

	app "github.com/isaqueveras/powersso/application/auth"
	domain "github.com/isaqueveras/powersso/domain/auth"
	"github.com/isaqueveras/powersso/i18n"
	"github.com/isaqueveras/powersso/oops"
	"github.com/isaqueveras/powersso/utils"
)

// @Router /v1/auth/create_account [POST]
func createAccount(ctx *gin.Context) {
	var input domain.CreateAccount
	if err := ctx.ShouldBindJSON(&input); err != nil {
		oops.Handling(ctx, err)
		return
	}

	url, err := app.CreateAccount(ctx, &input)
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"url":          *url,
		"message":      i18n.Value("create_account.message"),
		"instructions": i18n.Value("create_account.instructions"),
	})
}

// @Router /v1/auth/activation/{token} [POST]
func activation(ctx *gin.Context) {
	token, err := uuid.Parse(ctx.Param("token"))
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, map[string]string{
		"url":          *url,
		"message":      i18n.Value("create_account.message"),
		"instructions": i18n.Value("create_account.instructions"),
	})
}

// @Router /v1/auth/login [POST]
func login(ctx *gin.Context) {
	var input domain.Login
	if err := ctx.ShouldBindJSON(&input); err != nil {
		oops.Handling(ctx, err)
		return
	}

	input.ClientIP = utils.Pointer(ctx.ClientIP())
	input.UserAgent = utils.Pointer(ctx.Request.UserAgent())
	input.Validate()

	output, err := app.Login(ctx, &input)
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, output)
}

// @Router /v1/auth/change_password [put]
func changePassword(ctx *gin.Context) {
	in := &domain.ChangePassword{}
	if err := ctx.ShouldBindJSON(in); err != nil {
		oops.Handling(ctx, err)
		return
	}

	if ok := in.ValidatePassword(); !ok {
		oops.Handling(ctx, oops.New("Invalid passwords"))
		return
	}

	if err := app.ChangePassword(ctx, in); err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

// @Router /v1/auth/logout [DELETE]
func logout(ctx *gin.Context) {
	sessionID, err := uuid.Parse(gopowersso.GetSession(ctx).SessionID)
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	if err := app.Logout(ctx, utils.Pointer(sessionID)); err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusNoContent, utils.NoContent{})
}

// @Router /v1/auth/login/steps [GET]
func loginSteps(ctx *gin.Context) {
	res, err := app.LoginSteps(ctx, utils.Pointer(ctx.Query("email")))
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, res)
}

// @Router /v1/auth/user/{user_uuid}/disable [PUT]
func disable(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_uuid"))
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	if err = app.DisableUser(ctx, &userID); err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, utils.NoContent{})
}

// @Router /v1/auth/user/{user_uuid}/otp/configure [POST]
func configure(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_uuid"))
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	if err = app.Configure2FA(ctx, &userID); err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, utils.NoContent{})
}

// @Router /v1/auth/user/{user_uuid}/otp/unconfigure [PUT]
func unconfigure(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_uuid"))
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	if err = app.Unconfigure2FA(ctx, &userID); err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, utils.NoContent{})
}

// @Router /v1/auth/user/{user_uuid}/otp/qrcode [GET]
func qrcode(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_uuid"))
	if err != nil {
		oops.Handling(ctx, err)
		return
	}

	var url *string
	if url, err = app.GetQRCode2FA(ctx, &userID); err != nil {
		oops.Handling(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, map[string]*string{"url": url})
}
