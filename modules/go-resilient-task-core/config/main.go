package config

import (
	"database/sql"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/dao"
	handlerregistry "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/handler_registry"
	taskexecutor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_executor"
	taskresumer "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_resumer"
	taskservice "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_service"
	taskexecutiontrigger "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/triggering/task_execution_trigger"
	taskproperties "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task_properties"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type GoResilientTaskConfig struct {
	logger               *zap.Logger
	taskDao              dao.ITaskDao
	taskExecutor         taskexecutor.ITaskExecutor
	taskExecutionTrigger taskexecutiontrigger.ITasksExecutionTriggerer
	taskService          taskservice.ITasksService
	taskResumer          taskresumer.ITaskResumer
	taskHandlerRegistry  handlerregistry.ITaskHandlerRegistry
	taskProperties       taskproperties.ITaskProperties
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

func WithDefaultTaskRegistry() GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskHandlerRegistry = handlerregistry.NewTaskHandlerRegistry(
			config.requireLogger(),
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
		config.taskService = taskservice.NewTasksService(config.requireTaskDao(), config.requireTaskExecutionTrigger(), config.requireLogger(), config.requireTaskProperties())
	}
}

func WithKafkaTaskExecutionTrigger(kafkaClient *kgo.Client) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskExecutionTrigger = taskexecutiontrigger.NewKafkaTaskExecutionTrigger(config.requireTaskHandlerRegistry(), config.requireLogger(), kafkaClient, config.requireTaskExecutor(), config.requireTaskDao(), config.requireTaskProperties())
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
		config.taskResumer = taskresumer.NewTaskResumer(config.requireLogger(), config.requireTaskDao(), config.requireTaskExecutionTrigger(), config.requireTaskExecutor(), config.requireTaskHandlerRegistry(), config.requireTaskProperties())
	}
}

func WithTaskProperties(taskProperties taskproperties.ITaskProperties) GoResilientTaskConfigOption {
	return func(config *GoResilientTaskConfig) {
		config.taskProperties = taskProperties
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

	if config.taskProperties == nil {
		panic("TaskProperties is required")
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

func (c GoResilientTaskConfig) GetHandlerRegistry() handlerregistry.ITaskHandlerRegistry {
	return c.taskHandlerRegistry
}

func (c *GoResilientTaskConfig) GetTaskExecutor() taskexecutor.ITaskExecutor {
	return c.taskExecutor
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

func (c *GoResilientTaskConfig) requireTaskProperties() taskproperties.ITaskProperties {
	if c.taskProperties == nil {
		panic("TaskProperties is required")
	}
	return c.taskProperties
}
