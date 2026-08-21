package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"gojo/internal/app/ecode"
	"gojo/internal/app/response"
)

// RequestBodyLimit rejects declared oversized requests and caps streamed or
// chunked bodies before any handler attempts to decode them.
func RequestBodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 {
			c.Next()
			return
		}
		if c.Request.ContentLength > maxBytes {
			response.FailWithMessage(c, http.StatusRequestEntityTooLarge, ecode.InvalidParams, fmt.Sprintf("request body must not exceed %d bytes", maxBytes))
			c.Abort()
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
