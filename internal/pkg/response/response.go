// Package response 统一 API 信封与错误响应，避免对外泄漏内部错误细节。
package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

const CodeOK = "0000"
const CodeUnauthorized = "4010"
const CodeBadRequest = "4000"
const CodeNotFound = "4040"
const CodeServerError = "5000"

// R 对外统一响应体。
type R struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, R{Code: CodeOK, Message: "ok", Data: data})
}

func Fail(c *gin.Context, httpStatus int, code, message string) {
	c.JSON(httpStatus, R{Code: code, Message: message, Data: nil})
}

func BadRequest(c *gin.Context, message string) {
	Fail(c, http.StatusBadRequest, CodeBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, CodeUnauthorized, message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, CodeNotFound, message)
}

func ServerError(c *gin.Context, message string) {
	Fail(c, http.StatusInternalServerError, CodeServerError, message)
}

// ServerErrorLog 记录内部错误，对外只返回固定文案。
func ServerErrorLog(c *gin.Context, err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
	ServerError(c, message)
}

// BadRequestLog 记录内部错误，对外返回固定业务文案（用于需脱敏的失败）。
func BadRequestLog(c *gin.Context, err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
	BadRequest(c, message)
}

// NotFoundLog 记录内部错误，对外返回固定未找到文案。
func NotFoundLog(c *gin.Context, err error, message string) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	}
	NotFound(c, message)
}
