package main

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/config"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_handler"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func NewLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func NewResilientTaskConfiguration(logger *zap.Logger, taskHandler taskhandler.ITaskHandler, postgresClient *sql.DB, kafkaClient *kgo.Client) *config.GoResilientTaskConfig {

	configuration := config.NewGoResilientTaskConfig(
		config.WithLogger(logger),
		config.WithTaskHandlers([]taskhandler.ITaskHandler{taskHandler}),
		config.WithDefaultTaskRegistry(),
		config.WithPostgresTaskDao(postgresClient),
		config.WithDefaultTaskExecutor(),
		config.WithKafkaTaskExecutionTrigger(kafkaClient),
		config.WithDefaultTaskService(),
		config.WithDefaultTaskResumer(),
	)

	configuration.GetTaskResumer().ResumeProcessing()
	err := configuration.GetTaskExecutionTrigger().StartTasksProcessing()

	if err != nil {
		logger.Error("Failed to start task processing", zap.Error(err))
	}

	return configuration
}

func NewTaskHandler1(logger *zap.Logger) taskhandler.ITaskHandler {
	taskProcessor1 := taskhandler.NewTaskProcessor(logger, func(task taskmodel.ITask) {
		logger.Debug("Handler 1 processing task")
	})
	return taskhandler.NewTaskHandler(taskProcessor1, logger, func(task taskmodel.IBaseTask) bool {
		return task.GetType() == "test"
	})
}

var kafkaClientInstance *kgo.Client
var once sync.Once

func NewKgoClient(logger *zap.Logger) *kgo.Client {
	once.Do(func() {
		host := "localhost"
		port := "60277"

		// Construct the Kafka bootstrap server address
		bootstrapServer := fmt.Sprintf("%s:%s", host, port)

		seeds := []string{bootstrapServer}
		kafkaClient, err := kgo.NewClient(
			kgo.SeedBrokers(seeds...),
			kgo.AllowAutoTopicCreation(),
			kgo.ConsumerGroup("my-group-identifier12"),
			kgo.ConsumeTopics("tasks"))

		if err != nil {
			logger.Error("Failed to create Kafka client", zap.Error(err))
		}

		kafkaClientInstance = kafkaClient
	})

	return kafkaClientInstance
}

func NewPostgresClient(logger *zap.Logger) *sql.DB {
	// Construct the connection string
	connectionString := "postgres://user:password@localhost:5432/public?sslmode=disable"

	postgresClient, err := sql.Open("pgx", connectionString)
	if err != nil {
		logger.Error("Failed to create Postgres client", zap.Error(err))
		return nil
	}

	return postgresClient
}
