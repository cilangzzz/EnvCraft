package common

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Response 基础响应结构
type Response struct {
	Code    int         `json:"code"`           // 业务状态码
	Message string      `json:"message"`        // 提示信息
	Data    interface{} `json:"data,omitempty"` // 返回数据
	Meta    interface{} `json:"meta,omitempty"` // 元数据，可用于分页等扩展
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
	})
}

// Fail 返回500服务器错误
func Fail(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Code:    http.StatusInternalServerError,
		Message: message,
	})
}

// SuccessWithMeta 带元数据的成功响应
func SuccessWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    http.StatusOK,
		Message: "success",
		Data:    data,
		Meta:    meta,
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// Custom 自定义响应
func Custom(c *gin.Context, httpStatus, code int, message string, data interface{}) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// PageResponse 分页响应结构
type PageResponse struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// PageSuccess 分页成功响应
func PageSuccess(c *gin.Context, list interface{}, total int64, page int, pageSize int) {
	pageData := PageResponse{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	SuccessWithMeta(c, list, map[string]interface{}{
		"data":     pageData,
		"page":     page,
		"pageSize": pageSize,
		"total":    total,
	})
}
