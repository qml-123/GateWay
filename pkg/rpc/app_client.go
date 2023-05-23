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
	cli.app_service_client, err = appservice.NewClient(model.AppServiceName, option,
		client.WithConnectTimeout(5*time.Minute),
		client.WithRPCTimeout(15*time.Minute),
		client.WithGRPCConnPoolSize(20),
	)
	if err != nil {
		log.Println("NewClient log filed")
		return err
	}
	cli.m = map[string]func(c context.Context, ctx *app.RequestContext){
		"Ping":             cli.PingHandler,
		"Login":            cli.LoginHandler,
		"Register":         cli.RegisterHandler,
		"GetFileKey":       cli.GetFileKeyHandler,
		"Upload":           cli.UploadHandler,
		"GetFile":          cli.GetFileHandler,
		"GetFileChunkSize": cli.GetFileChunkSizeHandler,
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
		token, refreshToken, err = jwt.GenerateJWT(c, resp.GetUserId(), 5*time.Minute, true)
	} else {
		token, _, err = jwt.GenerateJWT(c, resp.GetUserId(), 5*time.Minute, true)
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

func getInt32(v interface{}) (int32, bool) {
	float, ok := v.(float64)
	if !ok {
		return 0, false
	}
	if i := int32(float); float64(i) == float {
		return i, true
	}
	return 0, false
}

func (cli *appClient) UploadHandler(c context.Context, ctx *app.RequestContext) {
	userID, ok := c.Value(model.UserID).(int64)
	if !ok {
		logger.Warn(c, "[GetFileKeyHandler] no userID found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}
	data, ok := c.Value(model.Data).(map[string]interface{})
	if !ok {
		logger.Warn(c, "[GetFileKeyHandler] no req data found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}

	//logger.Info(c, "upload, req: %v", data)

	isErr := false
	if _, ok = data["file"]; !ok {
		isErr = true
	}
	if _, ok = data["chunk_num"]; !ok {
		isErr = true
	}
	if _, ok = data["chunk_size"]; !ok {
		isErr = true
	}
	if _, ok = data["has_more"]; !ok {
		isErr = true
	}
	if _, ok = data["file_key"]; !ok {
		isErr = true
	}
	if _, ok = data["file_type"]; !ok {
		isErr = true
	}
	if isErr {
		utils.ErrorJSON(c, ctx, error_code.InvalidParam)
		return
	}

	var chunkNum, chunkSize int32
	if chunkNum, ok = getInt32(data["chunk_num"]); !ok {
		utils.ErrorJSON(c, ctx, error_code.NewStatus(error_code.InvalidParam.Code, "chunk_num is not int32"))
		return
	}
	if chunkSize, ok = getInt32(data["chunk_size"]); !ok {
		utils.ErrorJSON(c, ctx, error_code.NewStatus(error_code.InvalidParam.Code, "chunk_size is not int32"))
		return
	}
	req := &app_service.UploadFileRequest{
		UserId:    userID,
		File:      []byte(data["file"].(string)),
		FileKey:   data["file_key"].(string),
		FileType:  data["file_type"].(string),
		HasMore:   data["has_more"].(bool),
		ChunkNum:  int32(chunkNum),
		ChunkSize: int32(chunkSize),
	}
	//err := file.Upload(c, req.UserId, req.File, req.ChunkNum, req.ChunkSize, req.FileKey, req.HasMore, req.FileType)
	//if err != nil {
	//	logger.Error(c, "[Upload] failed, userID: %d, file_key: %s, chunk_num: %v, err: %v", userID, data["file_key"], data["chunk_num"], err)
	//	utils.ErrorJSON(c, ctx, error_code.InternalError)
	//	return
	//}
	resp, err := cli.app_service_client.Upload(c, req)
	if err != nil {
		logger.Error(c, "[Upload] failed, userID: %d, file_key: %s, chunk_num: %v, err: %v", userID, data["file_key"], data["chunk_num"], err)
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}
	if resp.GetBaseData().GetCode() != 0 {
		logger.Warn(c, "[Upload] resp code != 0, userId: %v, file_key: %v, chunk_num: %v, code: %v, message: %v", userID, data["file_key"], data["chunk_num"], resp.GetBaseData().GetCode(), resp.GetBaseData().GetMessage())
		utils.ErrorJSON(c, ctx, error_code.NewStatus(int(resp.GetBaseData().GetCode()), resp.GetBaseData().GetMessage()))
		return
	}
	utils.JSON(c, ctx, map[string]interface{}{})
}

func (cli *appClient) GetFileHandler(c context.Context, ctx *app.RequestContext) {
	userID, ok := c.Value(model.UserID).(int64)
	if !ok {
		logger.Warn(c, "[GetFileHandler] no userID found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}
	data, ok := c.Value(model.Data).(map[string]interface{})
	if !ok {
		logger.Warn(c, "[GetFileHandler] no req data found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}

	isErr := false
	if _, ok = data["chunk_num"]; !ok {
		isErr = true
	}
	if _, ok = data["file_key"]; !ok {
		isErr = true
	}
	if _, ok = data["file_key"].(string); !ok {
		isErr = true
	}
	if isErr {
		utils.ErrorJSON(c, ctx, error_code.InvalidParam)
		return
	}

	var chunkNum int32
	if chunkNum, ok = getInt32(data["chunk_num"]); !ok {
		utils.ErrorJSON(c, ctx, error_code.NewStatus(error_code.InvalidParam.Code, "chunk_num is not int32"))
		return
	}
	req := &app_service.GetFileRequest{
		UserId:   userID,
		FileKey:  data["file_key"].(string),
		ChunkNum: chunkNum,
	}
	resp, err := cli.app_service_client.GetFile(c, req)
	if err != nil {
		logger.Error(c, "[GetFileHandler] failed, userID: %d, file_key: %s, chunk_num: %v, err: %v", userID, data["file_key"], data["chunk_num"], err)
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}
	if resp.GetBaseData().GetCode() != 0 {
		logger.Warn(c, "[GetFileHandler] resp code != 0, userId: %v, file_key: %v, chunk_num: %v, code: %v, message: %v", userID, data["file_key"], data["chunk_num"], resp.GetBaseData().GetCode(), resp.GetBaseData().GetMessage())
		utils.ErrorJSON(c, ctx, error_code.NewStatus(int(resp.GetBaseData().GetCode()), resp.GetBaseData().GetMessage()))
		return
	}
	utils.JSON(c, ctx, map[string]interface{}{
		"file":       string(resp.GetFile()),
		"file_type":  resp.GetFileType(),
		"has_more":   resp.GetHasMore(),
		"chunk_size": resp.GetChunkSize(),
	})
}

func (cli *appClient) GetFileChunkSizeHandler(c context.Context, ctx *app.RequestContext) {
	userID, ok := c.Value(model.UserID).(int64)
	if !ok {
		logger.Warn(c, "[GetFileKeyHandler] no userID found")
		utils.ErrorJSON(c, ctx, error_code.InternalError)
		return
	}
	file_key := ctx.Param("file_key")
	resp, err := cli.app_service_client.GetFileChunkSize(c, &app_service.GetFileChunkNumRequest{
		UserId:  userID,
		FileKey: file_key,
	})
	if err != nil {
		logger.Error(c, "call GetFileChunkSize failed, err: %v", err.Error())
		utils.ErrorJSON(c, ctx, err)
		return
	}
	if resp.GetBaseData().GetCode() != 0 {
		logger.Warn(c, "resp baseData code != 0, code: %d, message: %s", resp.GetBaseData().GetCode(), resp.GetBaseData().GetMessage())
		utils.ErrorJSON(c, ctx, error_code.NewStatus(int(resp.GetBaseData().GetCode()), resp.GetBaseData().GetMessage()))
		return
	}
	utils.JSON(c, ctx, map[string]interface{}{
		"total_num": resp.GetChunkNum(),
	})
}
