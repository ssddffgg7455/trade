package controller

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/ssddffgg7455/logger"
)

type BaseController struct{}

// BaseContext 解析 gin.Context 的数据到 context.Context
func (ctl *BaseController) BaseContext(c *gin.Context) context.Context {
	ctx := logger.CtxWithTraceId()
	for k, v := range c.Keys {
		ctx = context.WithValue(ctx, k, v)
	}
	return ctx
}

// respond 回包
func respond(ctx *gin.Context, httpCode int, data interface{}, err error) {
	if data == nil {
		data = gin.H{}
	}

	var resp Response
	if err == nil {
		resp = Response{
			Result: data,
		}
	} else {
		resp = Response{
			Error:  err,
			Result: data,
		}
	}

	ctx.JSON(httpCode, resp)
}

// Response 服务回包
type Response struct {
	Error  interface{} `json:"error"`
	Result interface{} `json:"result"`
}
