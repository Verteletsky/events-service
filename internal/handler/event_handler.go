package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/godev/events-service/internal/logger"
	"github.com/godev/events-service/internal/model"
	"github.com/godev/events-service/internal/queue"
	"github.com/godev/events-service/internal/service"
)

type EventHandler struct {
	service service.IEventService
	queue   *queue.EventQueue
	log     *zap.Logger
}

func NewEventHandler(service service.IEventService, queue *queue.EventQueue) *EventHandler {
	return &EventHandler{
		service: service,
		queue:   queue,
		log:     logger.Get(),
	}
}

func (h *EventHandler) ListEvents(c *gin.Context) {
	offset, err := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "invalid offset parameter",
		})
		return
	}
	limit, err := strconv.ParseInt(c.DefaultQuery("limit", "100"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Message: "invalid limit parameter",
		})
		return
	}
	if limit > 100 {
		limit = 100
	}

	eventType := c.Query("type")

	h.log.Info("Listing events",
		zap.String("type", eventType),
		zap.Int64("offset", offset),
		zap.Int64("limit", limit))

	events, err := h.service.ListEvents(c.Request.Context(), eventType, offset, limit)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidEventType):
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		case errors.Is(err, service.ErrInvalidLimit):
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		default:
			h.log.Error("Failed to list events", zap.Error(err))
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to list events"})
		}
		return
	}

	c.JSON(http.StatusOK, model.EventsResponse{Events: events})
}

func (h *EventHandler) StartEvent(c *gin.Context) {
	var req model.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid request body"})
		return
	}

	if req.Type == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Event type is required"})
		return
	}

	h.log.Info("Starting event", zap.String("type", req.Type))

	err := h.queue.ProcessEvent(c.Request.Context(), req.Type, "start")
	if err != nil {
		h.log.Error("Failed to start event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to start event"})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *EventHandler) FinishEvent(c *gin.Context) {
	var req model.EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Invalid request body"})
		return
	}

	if req.Type == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: "Event type is required"})
		return
	}

	h.log.Info("Finishing event", zap.String("type", req.Type))

	err := h.queue.ProcessEvent(c.Request.Context(), req.Type, "finish")
	if err != nil {
		h.log.Error("Failed to finish event", zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Failed to finish event"})
		return
	}

	c.Status(http.StatusAccepted)
}
