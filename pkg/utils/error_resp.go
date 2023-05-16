package utils

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/qml-123/GateWay/model"
	"github.com/qml-123/app_log/error_code"
	"github.com/qml-123/app_log/logger"
)

func ErrorJSON(c context.Context, ctx *app.RequestContext, err error) {
	bizErr, ok := err.(*error_code.StatusError)
	logger.Info(c, "err: %v", err.Error())
	if !ok || errors.Is(err, error_code.InternalError) {
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	ctx.JSON(http.StatusOK, bizErr)
}

func JSON(c context.Context, ctx *app.RequestContext, data map[string]interface{}, tokens ...string) {
	if len(tokens) > 0 {
		ctx.SetCookie(
			model.JwtToken,
			tokens[0],
			int(5*time.Minute/time.Second),
			"/",
			"",
			protocol.CookieSameSiteStrictMode,
			false,
			true,
		)
	}
	if len(tokens) > 1 {
		ctx.SetCookie(
			model.RefreshToken,
			tokens[1],
			int(24*time.Hour/time.Second),
			"/",
			"",
			protocol.CookieSameSiteStrictMode,
			false,
			true,
		)
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    data,
	})
}
