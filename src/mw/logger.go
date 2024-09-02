package mw

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/logger/accesslog"
)

func NewAccessLog(h *server.Hertz) {
	format := "${ip} ${protocol} ${status} ${method} ${url} ${bytesReceived} ${bytesSent}"
	log := accesslog.New(accesslog.WithFormat(format))
	h.Use(log)
}
