package config

import (
	"log"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI      string
	Database string
	Options  *options.ClientOptions
}

type ServerConfig struct {
	Port     int
	LogLevel string
}

type Config struct {
	Mongo  MongoConfig
	Server ServerConfig
}

func New() *Config {
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017").(string)
	mongoDB := getEnv("MONGODB_DATABASE", "events").(string)
	serverPort := getEnv("SERVER_PORT", 8080).(int)
	logLevel := getEnv("LOG_LEVEL", "info").(string)

	clientOptions := options.Client().
		ApplyURI(mongoURI).
		SetConnectTimeout(10 * time.Second).
		SetServerSelectionTimeout(10 * time.Second).
		SetMaxPoolSize(100)

	return &Config{
		Mongo: MongoConfig{
			URI:      mongoURI,
			Database: mongoDB,
			Options:  clientOptions,
		},
		Server: ServerConfig{
			Port:     serverPort,
			LogLevel: logLevel,
		},
	}
}

func getEnv(key string, def interface{}) interface{} {
	switch defTyped := def.(type) {
	case string:
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return defTyped
	case int:
		if val, ok := os.LookupEnv(key); ok {
			intVal, err := strconv.Atoi(val)
			if err != nil {
				log.Printf("Warning: could not parse %s as int, using default %d", key, defTyped)
				return defTyped
			}
			return intVal
		}
		return defTyped
	default:
		return def
	}
}
