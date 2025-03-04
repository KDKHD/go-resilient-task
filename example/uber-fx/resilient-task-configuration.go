package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/config"
	concurrencypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/retry_policy"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_handler"
	taskhandleradapter "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_handler_adapter"
	taskprocessor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_processor"
	taskservice "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_service"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
	taskproperties "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task_properties"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func NewResilientTaskConfiguration(logger *zap.Logger, postgresClient *sql.DB, kafkaClient *kgo.Client, taskProperties taskproperties.ITaskProperties) *config.GoResilientTaskConfig {

	configuration := config.NewGoResilientTaskConfig(
		config.WithTaskProperties(taskProperties),
		config.WithLogger(logger),
		config.WithDefaultTaskRegistry(),
		config.WithPostgresTaskDao(postgresClient),
		config.WithDefaultTaskExecutor(),
		config.WithKafkaTaskExecutionTrigger(kafkaClient),
		config.WithDefaultTaskService(),
		config.WithDefaultTaskResumer(time.Second*2),
	)

	return configuration
}

func PaymentInitiatedHandler(logger *zap.Logger, config *config.GoResilientTaskConfig) taskhandler.ITaskHandler {
	return taskhandleradapter.NewTaskHandlerAdapterBuilder(
		func(task taskmodel.IBaseTask) bool {
			return task.GetType() == "payment_initiated"
		},
		taskprocessor.TaskProcessorFunc(
			func(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {
				logger.Debug("processing task payment initaiated", zap.String("task_data", task.GetData()))
				logger.Debug("Triggering task payment_processed", zap.String("task_data", task.GetData()))
				config.GetTaskService().AddTask(taskservice.AddTaskRequest{Type: "payment_processed", TaskId: uuid.New(), Data: []byte(task.GetData()), RunAfterTime: time.Now().UTC().Add(time.Second * 10), ExpectedQueueTime: time.Second * 120})
				return taskprocessor.ProcessResult{
					ResultCode: taskprocessor.DONE,
				}, nil
			},
		),
	).
		WithConcurrencyPolicy(
			concurrencypolicy.NewSimpleTaskConcurrencyPolicy(20, logger),
		).
		WithProcessingPolicy(
			processingpolicy.NewSimpleTaskProcessingPolicy(time.Minute * 1),
		).
		WithRetryPolicy(
			retrypolicy.NewExponentialRetryPolicy(
				retrypolicy.WithDelay(5*time.Second),
				retrypolicy.WithMultiplier(4),
				retrypolicy.WithMaxCount(3),
				retrypolicy.WithMaxDelay(20*time.Minute),
			),
		).
		Build()
}

func PaymentProcessedHandler(logger *zap.Logger) taskhandler.ITaskHandler {
	return taskhandleradapter.
		NewTaskHandlerAdapterBuilder(
			func(task taskmodel.IBaseTask) bool {
				return task.GetType() == "payment_processed"
			},
			taskprocessor.TaskProcessorFunc(
				func(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {
					logger.Debug("processing task payment_processed", zap.String("task_data", task.GetData()))
					return taskprocessor.ProcessResult{
						ResultCode: taskprocessor.DONE,
					}, nil
				},
			),
		).
		WithConcurrencyPolicy(
			concurrencypolicy.NewSimpleTaskConcurrencyPolicy(20, logger),
		).
		WithProcessingPolicy(
			processingpolicy.NewSimpleTaskProcessingPolicy(time.Minute * 1),
		).
		WithRetryPolicy(
			retrypolicy.NewExponentialRetryPolicy(
				retrypolicy.WithDelay(5*time.Second),
				retrypolicy.WithMultiplier(4),
				retrypolicy.WithMaxCount(3),
				retrypolicy.WithMaxDelay(20*time.Minute),
			),
		).
		Build()
}

func NewKgoClient(logger *zap.Logger) *kgo.Client {
	host := "localhost"
	port := "54944"

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

	return kafkaClient
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

func NewTaskProperties() taskproperties.ITaskProperties {
	return taskproperties.NewTaskProperties(taskproperties.WithTaskStuckTimeout(time.Minute * 5))
}

type LaunchResilientTaskFxInvokeParams struct {
	fx.In
	Configuration *config.GoResilientTaskConfig
	Handlers      []taskhandler.ITaskHandler `group:"handlers"`
	Logger        *zap.Logger
}

func InitiateResilientTaskFx() fx.Option {
	return fx.Invoke(func(params LaunchResilientTaskFxInvokeParams) {
		params.Configuration.GetHandlerRegistry().SetTaskHandlers(params.Handlers)
		params.Configuration.GetTaskExecutor().StartProcessing()
		params.Configuration.GetTaskResumer().ResumeProcessing()
		err := params.Configuration.GetTaskExecutionTrigger().StartTasksProcessing()

		if err != nil {
			params.Logger.Error("Failed to start task processing", zap.Error(err))
		}
	})
}
