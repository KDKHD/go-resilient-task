package config

import (
	"database/sql"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/dao"
	handlerregistry "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/handler_registry"
	taskexecutor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_executor"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_handler"
	taskresumer "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_resumer"
	taskservice "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_service"
	taskexecutiontrigger "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/triggering/task_execution_trigger"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type GoResilientTaskConfig struct {
	logger               *zap.Logger
	taskDao              dao.ITaskDao
	handlers             []taskhandler.ITaskHandler
	taskExecutor         taskexecutor.ITaskExecutor
	taskExecutionTrigger taskexecutiontrigger.ITasksExecutionTriggerer
	taskService          taskservice.ITasksService
	taskResumer          taskresumer.ITaskResumer
	taskHandlerRegistry  handlerregistry.ITaskHandlerRegistry
}

type GoResilientTaskConfigOption func(*GoResilientTaskConfig)

func WithLogger(logger *zap.Logger) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.logger = logger
	}
}

func WithTaskDao(taskDao dao.ITaskDao) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskDao = taskDao
	}
}

func WithPostgresTaskDao(postgresClient *sql.DB) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		if config.logger == nil {
			panic("Logger must be set before setting postgres task dao")
		}
		client, err := dao.NewPostgresTaskDao(postgresClient, config.requireLogger())
		if err != nil {
			panic(err)
		}
		config.taskDao = client
	}
}

func WithTaskHandlers(handlers []taskhandler.ITaskHandler) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.handlers = handlers
	}
}

func WithDefaultTaskRegistry() GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		taskHandlerRegistryOptions := []func(*handlerregistry.TaskHandlerRegistryConfig){
			handlerregistry.WithLogger(config.requireLogger()),
		}

		for _, handler := range config.requireHandlers() {
			taskHandlerRegistryOptions = append(taskHandlerRegistryOptions, handlerregistry.WithHandler(handler))
		}

		config.taskHandlerRegistry = handlerregistry.NewTaskHandlerRegistry(
			taskHandlerRegistryOptions...,
		)
	}
}

func WithTaskExecutor(taskExecutor taskexecutor.ITaskExecutor) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskExecutor = taskExecutor
	}
}

func WithTaskExecutionTrigger(taskExecutionTrigger taskexecutiontrigger.ITasksExecutionTriggerer) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskExecutionTrigger = taskExecutionTrigger
	}
}

func WithDefaultTaskExecutor() GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskExecutor = taskexecutor.NewTaskExecutor(config.requireTaskHandlerRegistry(), config.requireTaskDao(), config.requireLogger())
	}
}

func WithDefaultTaskService() GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskService = taskservice.NewTasksService(config.requireTaskDao(), config.requireTaskExecutionTrigger(), config.requireLogger())
	}
}

func WithKafkaTaskExecutionTrigger(kafkaClient *kgo.Client) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskExecutionTrigger = taskexecutiontrigger.NewKafkaTaskExecutionTrigger(config.requireTaskHandlerRegistry(), config.requireLogger(), kafkaClient, config.requireTaskExecutor(), config.requireTaskDao())
	}
}

func WithTasksService(taskService taskservice.ITasksService) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskService = taskService
	}
}

func WithTaskResumer(taskResumer taskresumer.ITaskResumer) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskResumer = taskResumer
	}
}

func WithTaskHandlerRegistry(taskHandlerRegistry handlerregistry.ITaskHandlerRegistry) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskHandlerRegistry = taskHandlerRegistry
	}
}

func WithDefaultTaskResumer() GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskResumer = taskresumer.NewTaskResumer(config.requireLogger(), config.requireTaskDao(), config.requireTaskExecutionTrigger(), config.requireTaskExecutor(), time.Second*2, config.requireTaskHandlerRegistry())
	}
}

func NewGoResilientTaskConfig(options ...GoResilientTaskConfigOption) *GoResilientTaskConfig {
	config := &GoResilientTaskConfig{}
	for _, option := range options {
		option(config)
	}

	if config.logger == nil {
		panic("Logger is required")
	}

	if config.taskDao == nil {
		panic("TaskDao is required")
	}

	if config.handlers == nil {
		panic("Handlers are required")
	}

	if config.taskExecutor == nil {
		panic("TaskExecutor is required")
	}

	if config.taskExecutionTrigger == nil {
		panic("TaskExecutionTrigger is required")
	}

	if config.taskService == nil {
		panic("TaskService is required")
	}

	if config.taskResumer == nil {
		panic("TaskResumer is required")
	}

	if config.taskHandlerRegistry == nil {
		panic("TaskHandlerRegistry is required")
	}

	return config
}

func (c *GoResilientTaskConfig) GetTaskService() taskservice.ITasksService {
	return c.taskService
}

func (c *GoResilientTaskConfig) GetTaskExecutionTrigger() taskexecutiontrigger.ITasksExecutionTriggerer {
	return c.taskExecutionTrigger
}

func (c *GoResilientTaskConfig) GetTaskResumer() taskresumer.ITaskResumer {
	return c.taskResumer
}

func (c *GoResilientTaskConfig) requireLogger() *zap.Logger {
	if c.logger == nil {
		panic("Logger is required")
	}
	return c.logger
}

func (c *GoResilientTaskConfig) requireTaskDao() dao.ITaskDao {
	if c.taskDao == nil {
		panic("TaskDao is required")
	}
	return c.taskDao
}

func (c *GoResilientTaskConfig) requireTaskHandlerRegistry() handlerregistry.ITaskHandlerRegistry {
	if c.taskHandlerRegistry == nil {
		panic("TaskHandlerRegistry is required")
	}
	return c.taskHandlerRegistry
}

func (c *GoResilientTaskConfig) requireTaskExecutor() taskexecutor.ITaskExecutor {
	if c.taskExecutor == nil {
		panic("TaskExecutor is required")
	}
	return c.taskExecutor
}

func (c *GoResilientTaskConfig) requireTaskExecutionTrigger() taskexecutiontrigger.ITasksExecutionTriggerer {
	if c.taskExecutionTrigger == nil {
		panic("TaskExecutionTrigger is required")
	}
	return c.taskExecutionTrigger
}

func (c *GoResilientTaskConfig) requireTaskService() taskservice.ITasksService {
	if c.taskService == nil {
		panic("TaskService is required")
	}
	return c.taskService
}

func (c *GoResilientTaskConfig) requireTaskResumer() taskresumer.ITaskResumer {
	if c.taskResumer == nil {
		panic("TaskResumer is required")
	}
	return c.taskResumer
}

func (c *GoResilientTaskConfig) requireHandlers() []taskhandler.ITaskHandler {
	if c.handlers == nil {
		panic("Handlers are required")
	}
	return c.handlers
}
