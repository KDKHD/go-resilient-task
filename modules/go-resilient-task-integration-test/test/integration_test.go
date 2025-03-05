package main

import (
	"context"
	"fmt"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/config"
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
	"github.com/stretchr/testify/assert"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func (s *IntegrationTestSuite) TestAddTask() {

	logger, _ := zap.NewDevelopment()
	defer logger.Sync() // Flushes buffer, if any

	kafkaClient, err := s.createKafkaClient(context.Background(),
		kgo.ConsumerGroup("default-consumer-group"),
		kgo.ConsumeTopics("tasks"))

	if err != nil {
		logger.Error("Failed to create kafka client", zap.Error(err))
	}

	taskHandler1 := taskhandleradapter.
		NewTaskHandlerAdapterBuilder(
			func(task taskmodel.IBaseTask) bool {
				return task.GetType() == "payment_initiated"
			},
			taskprocessor.TaskProcessorFunc(
				func(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {
					logger.Debug("processing task payment initaiated", zap.String("task_data", task.GetData()))
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

	taskHandler2 := taskhandleradapter.
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

	taskHandlers := []taskhandler.ITaskHandler{taskHandler1, taskHandler2}

	configuration := config.NewGoResilientTaskConfig(
		config.WithTaskProperties(
			taskproperties.NewTaskProperties(taskproperties.WithTaskStuckTimeout(time.Minute*5)),
		),
		config.WithLogger(logger),
		config.WithDefaultTaskRegistry(),
		config.WithPostgresTaskDao(s.postgresClient),
		config.WithDefaultTaskExecutor(),
		config.WithKafkaTaskExecutionTrigger(kafkaClient),
		config.WithDefaultTaskService(),
		config.WithDefaultTaskResumer(),
	)

	configuration.GetHandlerRegistry().SetTaskHandlers(taskHandlers)
	configuration.GetTaskExecutor().StartProcessing()
	configuration.GetTaskResumer().ResumeProcessing()
	configuration.GetTaskExecutionTrigger().StartTasksProcessing()

	for i := 0; i < 10; i++ {
		addTaskResponse1, error := configuration.GetTaskService().AddTask(taskservice.AddTaskRequest{Type: "payment_initiated", TaskId: uuid.New(), Data: []byte(fmt.Sprintf("TaskNumber %d", i)), RunAfterTime: time.Now().UTC().Add(time.Second * 5), ExpectedQueueTime: time.Second * 120})
		s.Require().NoError(error)

		assert.NotNil(s.T(), addTaskResponse1, "addTaskResponse1 should not be nil")
		assert.NotNil(s.T(), addTaskResponse1.TaskId.String(), "addTaskResponse1.TaskId should not be nil")

		addTaskResponse2, error := configuration.GetTaskService().AddTask(taskservice.AddTaskRequest{Type: "payment_processed", TaskId: uuid.New(), Data: []byte(fmt.Sprintf("TaskNumber %d", i)), RunAfterTime: time.Now().UTC().Add(time.Second * 10), ExpectedQueueTime: time.Second * 120})
		s.Require().NoError(error)
		assert.NotNil(s.T(), addTaskResponse2, "addTaskResponse2 should not be nil")
		assert.NotNil(s.T(), addTaskResponse2.TaskId.String(), "addTaskResponse2.TaskId should not be nil")

	}

	logger.Debug("Starting tasks processing")

	for {
		// sleep for 1 second
		time.Sleep(1 * time.Second)
		//tasksService.AddTask(task.AddTaskRequest{Type: "test2", TaskId: uuid.New(), Data: []byte(fmt.Sprintf("Surprise task"))})
	}
}
