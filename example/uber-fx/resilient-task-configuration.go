package main

import (
	"database/sql"
	"time"

	concurrencypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/concurrency_policy"
	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/processing_policy"
	retrypolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/retry_policy"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_handler"
	taskhandleradapter "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_handler_adapter"
	taskprocessor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_processor"
	taskservice "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_service"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
	taskproperties "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task_properties"
	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func PaymentInitiatedHandler(logger *zap.Logger, taskService taskservice.ITasksService) taskhandler.ITaskHandler {
	return taskhandleradapter.
		NewTaskHandlerAdapterBuilder(
			func(task taskmodel.IBaseTask) bool {
				return task.GetType() == "payment_initiated"
			},
			taskprocessor.TaskProcessorFunc(
				func(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {
					logger.Debug("processing task payment initaiated", zap.String("task_data", task.GetData()))
					// ... doing things ...
					logger.Debug("Triggering task payment_processed", zap.String("task_data", task.GetData()))
					taskService.AddTask(taskservice.AddTaskRequest{Type: "payment_processed", TaskId: uuid.New(), Data: []byte(task.GetData()), RunAfterTime: time.Now().UTC().Add(time.Second * 10), ExpectedQueueTime: time.Second * 120})
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

func NewKgoClient(logger *zap.Logger, taskProperties taskproperties.ITaskProperties) *kgo.Client {
	seeds := []string{taskProperties.GetKafka().GetBootstrapServers()}
	kafkaClient, err := kgo.NewClient(
		kgo.SeedBrokers(seeds...),
		kgo.AllowAutoTopicCreation(),
		kgo.ConsumerGroup(taskProperties.GetKafka().GetKafkaConsumerGroupId()),
		kgo.ConsumeTopics(taskProperties.GetKafka().GetKafkaTopicsNamespace()+"."+"tasks"))

	if err != nil {
		logger.Error("Failed to create Kafka client", zap.Error(err))
	}

	return kafkaClient
}

func NewPostgresClient(logger *zap.Logger) (*sql.DB, error) {
	connectionString := "postgres://user:password@localhost:5432/public?sslmode=disable"
	return sql.Open("pgx", connectionString)
}

func NewTaskProperties() taskproperties.ITaskProperties {
	return taskproperties.NewTaskProperties(
		taskproperties.WithTaskStuckTimeout(time.Minute*20),
		taskproperties.WithTaskResumer(
			taskproperties.NewTaskResumerProperties(
				taskproperties.WithBatchSize(1000),
				taskproperties.WithPollingInterval(time.Second*2),
				taskproperties.WithConcurrency(10),
			)),
		taskproperties.WithKafka(
			taskproperties.NewKafkaProperties(
				taskproperties.WithBootstrapServers("localhost:29092"),
				taskproperties.WithKafkaTopicsNamespace("my-group-identifier"),
				taskproperties.WithKafkaConsumerGroupId("tasks"),
			)),
	)
}
