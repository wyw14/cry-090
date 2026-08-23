package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wyw14/cry-090/internal/domain/common"
	"net/http"
)

func writeError(c *gin.Context, err error) {
	code := common.CodeOf(err)
	status := http.StatusInternalServerError
	switch code {
	case common.CodeInvalid:
		status = http.StatusBadRequest
	case common.CodeNotFound:
		status = http.StatusNotFound
	case common.CodeForbidden:
		status = http.StatusForbidden
	case common.CodeUnauthenticated:
		status = http.StatusUnauthorized
	case common.CodeConflict, common.CodeAlreadyProcessed, common.CodeVersionConflict:
		status = http.StatusConflict
	case common.CodeExpired:
		status = http.StatusGone
	case common.CodeCapacityFull:
		status = http.StatusUnprocessableEntity
	}
	body := gin.H{"code": string(code), "message": err.Error(), "request_id": c.GetString("request_id")}
	if domain, ok := err.(*common.DomainError); ok && len(domain.Fields) > 0 {
		body["fields"] = domain.Fields
	}
	c.AbortWithStatusJSON(status, body)
}
func actor(c *gin.Context) string { return c.GetHeader("X-User-ID") }
