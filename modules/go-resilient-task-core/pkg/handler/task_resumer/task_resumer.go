package taskresumer

import (
	"sync/atomic"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/dao"
	handlerregistry "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/handler_registry"
	processingpolicy "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/processing_policy"
	taskexecutor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_executor"
	taskexecutiontrigger "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/triggering/task_execution_trigger"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
	"go.uber.org/zap"
)

type ITaskResumer interface {
	PauseProcessing()
	ResumeProcessing()
}

type TaskResumer struct {
	logger                  *zap.Logger
	paused                  *atomic.Bool
	taskDao                 dao.ITaskDao
	tasksExecutionTriggerer taskexecutiontrigger.ITasksExecutionTriggerer
	taskExecutor            taskexecutor.ITaskExecutor
	taskHandlerRegistry     handlerregistry.ITaskHandlerRegistry
	pollingInterval         *time.Ticker
}

func NewTaskResumer(logger *zap.Logger, taskDao dao.ITaskDao, tasksExecutionTriggerer taskexecutiontrigger.ITasksExecutionTriggerer, taskExecutor taskexecutor.ITaskExecutor, pollingInterval time.Duration, taskHandlerRegistry handlerregistry.ITaskHandlerRegistry) *TaskResumer {
	taskResumer := TaskResumer{
		logger:                  logger,
		paused:                  &atomic.Bool{},
		taskDao:                 taskDao,
		tasksExecutionTriggerer: tasksExecutionTriggerer,
		taskExecutor:            taskExecutor,
		pollingInterval:         time.NewTicker(pollingInterval),
		taskHandlerRegistry:     taskHandlerRegistry,
	}

	return &taskResumer
}

func (tr *TaskResumer) PauseProcessing() {
	tr.paused.Store(true)
}

func (tr *TaskResumer) ResumeProcessing() {
	tr.paused.Store(false)
	tr.startProcessing()
}

func (tr *TaskResumer) startProcessing() {
	go tr.stuckTaskWorker()
}

func (tr *TaskResumer) stuckTaskWorker() {
	for range tr.pollingInterval.C {
		tr.logger.Debug("Polling stuck and waiting tasks")
		if tr.paused.Load() {
			tr.logger.Debug("Processing paused")
			return
		}
		tr.pickupStuckTasks()
		tr.pickupWaitingTasks()
	}
}

func (tr *TaskResumer) pickupWaitingTasks() {

	if tr.paused.Load() {
		tr.logger.Debug("Processing paused")
		return
	}

	stuckTasks, err := tr.taskDao.GetStuckTasks(100, taskmodel.WAITING)
	if err != nil {
		tr.logger.Error(err.Error())
		return
	}

	for _, task := range stuckTasks.StuckTasks {
		tr.handleWaitingTask(task)
	}
}

func (tr *TaskResumer) handleWaitingTask(stuckTask taskmodel.IStuckTask) {
	nextEventTime, err := tr.taskHandlerRegistry.GetExpectedProcessingMoment(stuckTask)

	if err != nil {
		tr.logger.Error(err.Error())
		return
	}

	submitted, err := tr.taskDao.MarkAsSubmitted(stuckTask.GetId(), stuckTask.GetVersion(), nextEventTime)
	if err != nil {
		tr.logger.Error(err.Error())
		return
	}

	if !submitted {
		tr.logger.Error("Not submitted")
		return
	}
	tr.tasksExecutionTriggerer.Trigger(stuckTask.IncrementVersion())
}

func (tr *TaskResumer) pickupStuckTasks() {
	statusesToCheck := []taskmodel.TaskStatus{taskmodel.NEW, taskmodel.SUBMITTED, taskmodel.PROCESSING}

	for _, status := range statusesToCheck {
		if tr.paused.Load() {
			tr.logger.Debug("Processing paused")
			return
		}

		stuckTasks, err := tr.taskDao.GetStuckTasks(100, status)
		if err != nil {
			tr.logger.Error(err.Error())
			continue
		}

		for _, task := range stuckTasks.StuckTasks {
			tr.handleStuckTask(task)
		}
	}
}

func (tr *TaskResumer) handleStuckTask(stuckTask taskmodel.IStuckTask) {
	taskHandler, err := tr.taskHandlerRegistry.GetTaskHandler(stuckTask)

	if err != nil {
		tr.logger.Error(err.Error())
	}

	var taskResolutionStrategy processingpolicy.StuckTaskResolutionStrategy

	if taskHandler == nil {
		tr.logger.Error("No handler found", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
	} else {
		taskProcessingStrategy := taskHandler.GetProcessingPolicy(stuckTask)
		if taskProcessingStrategy == nil {
			tr.logger.Error("No processing policy found for task", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		} else {
			taskResolutionStrategy = taskProcessingStrategy.GetStuckTaskResolutionStrategy(stuckTask)
		}
	}

	if taskmodel.PROCESSING != stuckTask.GetStatus() {
		tr.logger.Debug("Task is stuck", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		tr.retryTask(stuckTask)
		return
	}

	if taskResolutionStrategy == processingpolicy.NOOP {
		tr.logger.Error("No stuck task resolution strategy found", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		return
	}

	switch taskResolutionStrategy {
	case processingpolicy.RETRY:
		tr.logger.Debug("Retrying stuck task", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		tr.retryTask(stuckTask)
		return
	case processingpolicy.MARK_AS_ERROR:
		tr.logger.Debug("Marking task as error", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		tr.markAsError(stuckTask)
		return
	case processingpolicy.MARK_AS_FAILED:
		tr.logger.Debug("Marking task as failed", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		updated, err := tr.taskDao.SetStatus(stuckTask.GetId(), taskmodel.FAILED, stuckTask.GetVersion())
		if err != nil {
			tr.logger.Error(err.Error())
			return
		}

		if !updated {
			tr.logger.Error("Task version mismatch", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
			return
		}
		return
	case processingpolicy.IGNORE:
		tr.logger.Debug("Ignoring stuck task", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		return
	default:
		tr.logger.Error("Unknown task resolution strategy", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		return
	}
}

func (tr TaskResumer) markAsError(stuckTask taskmodel.IStuckTask) {
	updated, err := tr.taskDao.SetStatus(stuckTask.GetId(), taskmodel.ERROR, stuckTask.GetVersion())
	if err != nil {
		tr.logger.Error(err.Error())
		return
	}

	if !updated {
		tr.logger.Error("Failed to markAsError, task version mismatch", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		return
	}

	tr.logger.Debug("Marked as error", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
}

func (tr TaskResumer) retryTask(stuckTask taskmodel.IStuckTask) {

	nextEventTime, err := tr.taskHandlerRegistry.GetExpectedProcessingMoment(stuckTask)

	if err != nil {
		tr.logger.Error(err.Error())
		return
	}

	updated, err := tr.taskDao.MarkAsSubmitted(stuckTask.GetId(), stuckTask.GetVersion(), nextEventTime)
	if err != nil {
		tr.logger.Error(err.Error())
		return
	}

	if !updated {
		tr.logger.Error("Task version mismatch", zap.String("task", stuckTask.GetId().String()), zap.String("status", stuckTask.GetStatus().String()), zap.String("type", stuckTask.GetType()))
		return
	}

	task := stuckTask.IncrementVersion()
	tr.tasksExecutionTriggerer.Trigger(task)
}
