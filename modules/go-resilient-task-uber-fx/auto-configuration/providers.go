package autoconfiguration

import (
	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/dao"
	handlerregistry "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/handler_registry"
	taskexecutor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_executor"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_handler"
	taskresumer "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_resumer"
	taskservice "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_service"
	taskexecutiontrigger "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/triggering/task_execution_trigger"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Provider() fx.Option {
	return fx.Provide(
		zap.NewProduction,
		dao.NewPostgresTaskDao,
		handlerregistry.NewTaskHandlerRegistry,
		taskexecutor.NewTaskExecutor,
		taskservice.NewTasksService,
		taskexecutiontrigger.NewKafkaTaskExecutionTrigger,
		taskresumer.NewTaskResumer,
	)
}

type LaunchResilientTaskFxInvokeParams struct {
	fx.In
	HandlerRegistry      handlerregistry.ITaskHandlerRegistry
	TaskExecutor         taskexecutor.ITaskExecutor
	TaskResumer          taskresumer.ITaskResumer
	TaskExecutionTrigger taskexecutiontrigger.ITasksExecutionTriggerer
	Handlers             []taskhandler.ITaskHandler `group:"grt_handlers"`
	Logger               *zap.Logger
}

func InitiateResilientTaskFx() fx.Option {
	return fx.Invoke(func(params LaunchResilientTaskFxInvokeParams) {
		params.HandlerRegistry.SetTaskHandlers(params.Handlers)
		params.TaskExecutor.StartProcessing()
		params.TaskResumer.ResumeProcessing()
		err := params.TaskExecutionTrigger.StartTasksProcessing()

		if err != nil {
			params.Logger.Error("Failed to start task processing", zap.Error(err))
		}
	})
}

func AsHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(taskhandler.ITaskHandler)),
		fx.ResultTags(`group:"grt_handlers"`),
	)
}
