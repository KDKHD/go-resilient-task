package main

import (
	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/config"
	"go.uber.org/fx"
)

func main() {

	fx.New(
		fx.Provide(NewLogger),
		fx.Provide(NewTaskHandler1),
		fx.Provide(NewResilientTaskConfiguration),
		fx.Provide(NewKgoClient),
		fx.Provide(NewPostgresClient),
		fx.Invoke(func(*config.GoResilientTaskConfig) {}),
	).Run()

	/* logger, _ := zap.NewDevelopment()
	defer logger.Sync() // Flushes buffer, if any

	taskProcessor1 := taskhandler.NewTaskProcessor(logger, func(task taskmodel.ITask) {
		logger.Debug("Handler 1 processing task")
	})
	taskHandler1 := taskhandler.NewTaskHandler(taskProcessor1, logger, func(task taskmodel.IBaseTask) bool {
		return task.GetType() == "test"
	})
	taskProcessor2 := taskhandler.NewTaskProcessor(logger, func(task taskmodel.ITask) {
		logger.Debug("Handler 2 processing task")
	})
	taskHandler2 := taskhandler.NewTaskHandler(taskProcessor2, logger, func(task taskmodel.IBaseTask) bool {
		return task.GetType() == "test2"
	})

	taskHandlers := []taskhandler.ITaskHandler{taskHandler1, taskHandler2}

	configuration := config.NewGoResilientTaskConfig(
		config.WithLogger(logger),
		config.WithTaskHandlers(taskHandlers),
		config.WithDefaultTaskRegistry(),
		config.WithPostgresTaskDao(s.postgresClient),
		config.WithDefaultTaskExecutor(),
		config.WithKafkaTaskExecutionTrigger(kafkaClient),
		config.WithDefaultTaskService(),
		config.WithDefaultTaskResumer(),
	) */
}
