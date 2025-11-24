package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code int         `json:"code"`           // 业务状态码：1 成功，0 失败，其它表示错误
	Msg  string      `json:"msg"`            // 提示信息
	Data interface{} `json:"data,omitempty"` // 返回数据（可为空）
}

// Success 返回成功
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 1,
		Msg:  "success",
		Data: data,
	})
}

// Fail 返回业务失败（如参数错误、找不到数据）
func Fail(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: 0,
		Msg:  msg,
	})
}

// Error 返回系统错误（例如数据库错误）
func Error(c *gin.Context, err error) {
	c.JSON(http.StatusOK, Response{
		Code: -1,
		Msg:  err.Error(),
	})
}
