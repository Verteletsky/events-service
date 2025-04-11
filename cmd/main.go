package main

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/godev/events-service/internal/config"
	"github.com/godev/events-service/internal/db"
	"github.com/godev/events-service/internal/handler"
	"github.com/godev/events-service/internal/logger"
	"github.com/godev/events-service/internal/queue"
	mongorepo "github.com/godev/events-service/internal/repository/mongo"
	"github.com/godev/events-service/internal/service"
)

func main() {
	logger.Init()
	defer logger.Sync()
	log := logger.Get()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.New()

	mongodb, err := db.NewMongoDB(ctx, log, &cfg.Mongo)
	if err != nil {
		log.Fatal("Failed to initialize MongoDB", zap.Error(err))
	}
	defer func() {
		if err := mongodb.Close(); err != nil {
			log.Error("Failed to close MongoDB connection", zap.Error(err))
		}
	}()

	eventRepo := mongorepo.NewEventRepository(mongodb.GetDatabase())
	eventService := service.NewEventService(eventRepo)

	eventQueue := queue.NewEventQueue(ctx, eventService, log, 50, 500)
	defer eventQueue.Close()

	eventHandler := handler.NewEventHandler(eventService, eventQueue)

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.GET("", eventHandler.ListEvents)
		v1.POST("/start", eventHandler.StartEvent)
		v1.POST("/finish", eventHandler.FinishEvent)
	}

	RunServer(router, cfg, log)
}
