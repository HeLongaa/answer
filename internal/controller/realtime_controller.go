package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apache/answer/internal/base/middleware"
	"github.com/apache/answer/internal/service/realtime"
	"github.com/gin-gonic/gin"
)

type RealtimeController struct {
	realtimeService *realtime.Service
}

func NewRealtimeController(realtimeService *realtime.Service) *RealtimeController {
	return &RealtimeController{realtimeService: realtimeService}
}

func (c *RealtimeController) Events(ctx *gin.Context) {
	userID := middleware.GetLoginUserIDFromContext(ctx)
	events, unsubscribe := c.realtimeService.Subscribe(userID)
	defer unsubscribe()

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")
	ctx.Status(http.StatusOK)

	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()

	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Request.Context().Done():
			return
		case <-ticker.C:
			_, _ = fmt.Fprint(ctx.Writer, ": ping\n\n")
			flusher.Flush()
		case event := <-events:
			payload, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(ctx.Writer, "event: %s\n", event.Type)
			_, _ = fmt.Fprintf(ctx.Writer, "data: %s\n\n", payload)
			flusher.Flush()
		}
	}
}
