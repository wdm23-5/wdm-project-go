package common

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"time"
)

type ginWriterWrapper struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w ginWriterWrapper) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func GinLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		wrapped := &ginWriterWrapper{ResponseWriter: ctx.Writer, body: bytes.NewBuffer(nil)}
		ctx.Writer = wrapped
		start := time.Now()
		path := ctx.Request.URL.Path
		ctx.Next()
		latency := time.Now().Sub(start)
		status := ctx.Writer.Status()
		_, _ = fmt.Fprintf(
			gin.DefaultWriter, "st %v | lat %v | url [%v] %v | http %v | resp %v |\n",
			start.Format("15:04:05.000"), latency, ctx.Request.Method, path, status, wrapped.body,
		)
	}
}
