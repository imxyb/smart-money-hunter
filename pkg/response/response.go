package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"smart-money/pkg/errcode"
	"smart-money/pkg/log"
)

// BadRequest return bad request
func BadRequest(c *gin.Context, code int, err error) {
	baseErrResponse(c, http.StatusBadRequest, code, err)
}

// NotFound return not found
func NotFound(c *gin.Context, code int, err error) {
	baseErrResponse(c, http.StatusNotFound, code, err)
}

// Unauthorized 未认证错误
func Unauthorized(c *gin.Context, code int, err error) {
	baseErrResponse(c, http.StatusUnauthorized, code, err)
}

// InternalServerError 服务器错误
func InternalServerError(c *gin.Context, err error) {
	log.Errorf("internal server error:%s", err)
	baseErrResponse(c, http.StatusInternalServerError, errcode.InternalServerError, fmt.Errorf("internal server error"))
}

// OK return ok
func OK(c *gin.Context, data interface{}, message ...string) {
	msg := "ok"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(
		http.StatusOK, gin.H{
			"code":    0,
			"data":    data,
			"message": msg,
		},
	)
}

// OKList return ok list
func OKList(c *gin.Context, total int64, list interface{}, message ...string) {
	msg := "ok"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(
		http.StatusOK, gin.H{
			"code": 0,
			"data": gin.H{
				"total": total,
				"list":  list,
			},
			"message": msg,
		},
	)
}

func baseErrResponse(c *gin.Context, httpCode, code int, err error) {
	c.AbortWithStatusJSON(
		httpCode, gin.H{
			"code":    code,
			"message": err.Error(),
		},
	)
}
