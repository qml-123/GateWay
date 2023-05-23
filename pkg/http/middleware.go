package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/qml-123/GateWay/model"
	"github.com/qml-123/GateWay/pkg/id"
	"github.com/qml-123/GateWay/pkg/jwt"
	"github.com/qml-123/GateWay/pkg/utils"
	"github.com/qml-123/app_log/error_code"
)

// CORS utils
func CorsMiddleware() app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		var logID string
		if logID = string(ctx.GetHeader(model.LogID)); len(logID) == 0 {
			logID = id.GenerateIDBase58()
		}
		c = context.WithValue(c, model.LogID, logID)

		if string(ctx.Method()) == consts.MethodOptions {
			ctx.Status(http.StatusOK)
			return
		}

		if strings.Contains(string(ctx.Path()), "ping") {
			ctx.Next(c)
			return
		}

		if strings.Contains(string(ctx.Path()), "search") {
			ctx.Next(c)
			return
		}

		var m map[string]interface{}
		if string(ctx.Method()) == consts.MethodPost {
			if err := json.Unmarshal(ctx.Request.Body(), &m); err != nil {
				utils.ErrorJSON(c, ctx, error_code.InvalidParam)
				ctx.Abort()
				return
			}
		}

		c = context.WithValue(c, model.Data, m)

		isTokenNew := false
		var token string
		token = string(ctx.Cookie(model.JwtToken))
		if len(token) > 0 {
			c = context.WithValue(c, model.JwtToken, token)
		}

		if len(token) == 0 && !strings.Contains(string(ctx.Path()), "/login") && !strings.Contains(string(ctx.Path()), "/register") {
			utils.ErrorJSON(c, ctx, error_code.NewStatus(error_code.InvalidToken.Code, "no token"))
			ctx.Abort()
			return
		}
		if len(token) > 0 && !strings.Contains(string(ctx.Path()), "/login") {
			re_token := string(ctx.Cookie(model.RefreshToken))
			if re_token != "" {
				c = context.WithValue(c, model.RefreshToken, re_token)
			}
			var err error
			var userID int64
			userID, token, isTokenNew, err = jwt.CheckTokenExpiry(c, token)
			if err != nil {
				utils.ErrorJSON(c, ctx, error_code.InvalidToken)
				ctx.Abort()
				return
			}
			c = context.WithValue(c, model.UserID, userID)
		}

		ctx.Next(c)

		ctx.Header(model.LogID, logID)

		if isTokenNew {
			ctx.SetCookie(
				model.JwtToken,
				token,
				int(5*time.Minute/time.Second),
				"/",
				"",
				protocol.CookieSameSiteStrictMode,
				false,
				true,
			)
		}
	}
}
