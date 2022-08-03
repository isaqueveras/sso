package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/isaqueveras/lingo"
	"github.com/isaqueveras/power-sso/pkg/i18n"
)

// SetupI18n implements i18n configuration to be used in middleware
func SetupI18n(lang *lingo.L) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		i18n.Setup(ctx, lang)
		ctx.Next()
	}
}
