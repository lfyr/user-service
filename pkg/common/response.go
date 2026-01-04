package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// HTTP 统一响应格式
type ApiResponse struct {
	Code    int         `json:"code"`    // 响应码：0成功/1业务失败/2降级/500系统错误
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 响应数据（可选）
}

// 响应成功
func Success(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "操作成功"
	}
	c.JSON(http.StatusOK, ApiResponse{
		Code:    0,
		Message: message,
		Data:    data,
	})
}

// 响应失败（业务错误，如参数错误、用户名重复）
func Fail(c *gin.Context, code int, message string) {
	if code == 0 {
		code = 1
	}
	if message == "" {
		message = "操作失败"
	}
	c.JSON(http.StatusOK, ApiResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// 响应错误（系统错误，如微服务调用失败）
func Error(c *gin.Context, message string) {
	if message == "" {
		message = "系统内部错误"
	}
	c.JSON(http.StatusInternalServerError, ApiResponse{
		Code:    500,
		Message: message,
		Data:    nil,
	})
}

// 微服务响应转换为 HTTP 响应数据
func MicroRespToData(userId string) interface{} {
	if userId != "" {
		return map[string]string{"user_id": userId}
	}
	return nil
}
