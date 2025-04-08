package main

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/godev/events-service/internal/config"
	"github.com/godev/events-service/internal/db"
	"github.com/godev/events-service/internal/handler"
	"github.com/godev/events-service/internal/logger"
	mongorepo "github.com/godev/events-service/internal/repository/mongo"
	"github.com/godev/events-service/internal/service"
)

func main() {
	logger.Init()
	defer logger.Sync()
	log := logger.Get()

	cfg := config.New()

	mongodb, err := db.NewMongoDB(log, &cfg.Mongo)
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
	eventHandler := handler.NewEventHandler(eventService)

	router := gin.Default()

	v1 := router.Group("/v1")
	{
		v1.GET("", eventHandler.ListEvents)
		v1.POST("/start", eventHandler.StartEvent)
		v1.POST("/finish", eventHandler.FinishEvent)
	}

	RunServer(router, cfg.Server.Port, log)
}
