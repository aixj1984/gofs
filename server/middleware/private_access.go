package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/no-src/gofs/server/handler"
	"github.com/no-src/log"
	"net"
	"net/http"
)

type privateAccessHandler struct {
	logger log.Logger
}

func NewPrivateAccessHandler(logger log.Logger) handler.GinHandler {
	return &privateAccessHandler{
		logger: logger,
	}
}

func (h *privateAccessHandler) Handle(c *gin.Context) {
	ip := net.ParseIP(c.ClientIP())
	if !ip.IsPrivate() && !ip.IsLoopback() {
		h.logger.Warn("access deny, client ip is [%s], path is [%s]", c.ClientIP(), c.FullPath())
		c.Abort()
		c.String(http.StatusUnauthorized, "access deny")
	}
}