package rpc

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/andeya/ameda"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/client"
	goredis "github.com/go-redis/redis"
	"github.com/qml-123/GateWay/model"
	"github.com/qml-123/GateWay/pkg/jwt"
	"github.com/qml-123/GateWay/pkg/redis"
	"github.com/qml-123/GateWay/pkg/utils"
	"github.com/qml-123/app_log/error_code"
	app_service "github.com/qml-123/app_log/kitex_gen/app"
	"github.com/qml-123/app_log/kitex_gen/app/appservice"
	"github.com/qml-123/app_log/logger"
)

type appClient struct {
	app_service_client appservice.Client
	m                  map[string]func(c context.Context, ctx *app.RequestContext)
}

func newappClient() *appClient {
	return &appClient{
		m: make(map[string]func(c context.Context, ctx *app.RequestContext)),
	}
}

func (cli *appClient) initClient() error {
	option := client.WithResolver(r)
	var err error
	cli.app_service_client, err = appservice.NewClient(model.AppServiceName, option)
	if err != nil {
		log.Println("NewClient log filed")
		return err
	}
	cli.m = map[string]func(c context.Context, ctx *app.RequestContext){
		"Ping":       cli.PingHandler,
		"Login":      cli.LoginHandler,
		"Register":   cli.RegisterHandler,
		"GetFileKey": cli.GetFileKeyHandler,
	}
	return nil
}

func (cli *appClient) getHandler(method string) (func(c context.Context, ctx *app.RequestContext), error) {
	if _, ok := cli.m[method]; !ok {
		return nil, fmt.Errorf("no method")
	}
	return cli.m[method], nil
}

func (cli *appClient) PingHandler(c context.Context, ctx *app.RequestContext) {
	resp, err := cli.app_service_client.Ping(c, &app_service.PingRequest{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, error_code.InternalError)
		return
	}
	ctx.JSON(http.StatusOK, map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

func (cli *appClient) RegisterHandler(c context.Context, ctx *app.RequestContext) {
	data, ok := c.Value(model.Data).(map[string]interface{})
	if !ok {
		logger.Warn(c, "[RegisterHandler] no req data found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}

	req := &app_service.RegisteRequest{}
	_, ok1 := data["user_name"].(string)
	_, ok2 := data["pass_word"].(string)
	if data["user_name"] == "" || data["pass_word"] == "" || !ok1 || !ok2 {
		logger.Warn(c, "[RegisterHandler], no user_name or pass_word found")
		utils.ErrorJSON(c, ctx, error_code.InvalidParam)
		return
	}
	req.UserName = data["user_name"].(string)
	req.Password = data["pass_word"].(string)
	if _, ok = data["email"].(string); ok && data["email"] != "" {
		req.Email = ameda.StringToStringPtr(data["email"].(string))
	}
	if _, ok = data["phone_number"].(string); ok && data["phone_number"] != "" {
		req.PhoneNumber = ameda.StringToStringPtr(data["phone_number"].(string))
	}
	resp, err := cli.app_service_client.Register(c, req)
	if err != nil {
		logger.Error(c, "call Register failed, err: %v", err.Error())
		utils.ErrorJSON(c, ctx, err)
		return
	}
	if resp.GetBaseData().GetCode() != 0 {
		utils.ErrorJSON(c, ctx, error_code.NewStatus(int(resp.GetBaseData().GetCode()), resp.GetBaseData().GetMessage()))
		return
	}
	utils.JSON(c, ctx, map[string]interface{}{})
}

func (cli *appClient) LoginHandler(c context.Context, ctx *app.RequestContext) {
	data, ok := c.Value(model.Data).(map[string]interface{})
	if !ok {
		logger.Warn(c, "[LoginHandler] no req data found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}

	req := &app_service.LoginRequest{}
	user_name, ok1 := data["user_name"].(string)
	pass_word, ok2 := data["pass_word"].(string)
	if !ok1 || !ok2 || user_name == "" || pass_word == "" {
		utils.ErrorJSON(c, ctx, error_code.InvalidParam)
		return
	}
	if data["user_name"] != "" {
		req.UserName = ameda.StringToStringPtr(user_name)
	}
	if data["pass_word"] != "" {
		req.Password = ameda.StringToStringPtr(pass_word)
	}
	resp, err := cli.app_service_client.Login(c, req)
	if err != nil {
		logger.Error(c, "call Login failed, err: %v", err.Error())
		utils.ErrorJSON(c, ctx, err)
		return
	}
	var token, refreshToken string
	refreshToken, err = redis.Get(fmt.Sprintf("%d_refresh_token", resp.UserId))
	if err == goredis.Nil {
		token, refreshToken, err = jwt.GenerateJWT(resp.GetUserId(), 5*time.Minute, true)
	} else {
		token, _, err = jwt.GenerateJWT(resp.GetUserId(), 5*time.Minute, true)
	}

	if err != nil {
		logger.Warn(c, "gen jwt token error, err: %v", err)
		utils.ErrorJSON(c, ctx, err)
		return
	}
	if resp.GetBaseData().GetCode() != 0 {
		utils.ErrorJSON(c, ctx, error_code.NewStatus(int(resp.GetBaseData().GetCode()), resp.GetBaseData().GetMessage()))
		return
	}

	redis.Set(fmt.Sprintf("%d_refresh_token", resp.UserId), refreshToken, 24*time.Hour)

	utils.JSON(c, ctx, map[string]interface{}{
		"user_id":       fmt.Sprintf("%d", resp.UserId),
		"token":         token,
		"refresh_token": refreshToken,
	}, token, refreshToken)
}

func (cli *appClient) GetFileKeyHandler(c context.Context, ctx *app.RequestContext) {
	userID, ok := c.Value(model.UserID).(int64)
	if !ok {
		logger.Warn(c, "[GetFileKeyHandler] no userID found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}
	resp, err := cli.app_service_client.GetFileKey(c, &app_service.GetFileKeyRequest{
		UserId: userID,
	})
	if err != nil {
		logger.Error(c, "call GetFileKey failed, err: %v", err.Error())
		utils.ErrorJSON(c, ctx, err)
		return
	}
	if resp.GetBaseData().GetCode() != 0 {
		logger.Warn(c, "resp baseData code != 0, code: %d, message: %s", resp.GetBaseData().GetCode(), resp.GetBaseData().GetMessage())
		utils.ErrorJSON(c, ctx, error_code.NewStatus(int(resp.GetBaseData().GetCode()), resp.GetBaseData().GetMessage()))
		return
	}
	utils.JSON(c, ctx, map[string]interface{}{
		"file_key": resp.FileKey,
	})
}
