package taskexecutor

import (
	"errors"
	"sync"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/dao"
	handlerregistry "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/handler_registry"
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_handler"
	taskprocessor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/task_processor"
	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/handler/triggering"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task/pkg/model/task"
	"go.uber.org/zap"
)

type ITaskExecutor interface {
	SubmitTask(taskmodel.IBaseTask) error
}

type WorkerPoolProperties struct {
	taskTriggeringForProcessing chan triggering.TaskTriggering
	taskHandler                 taskhandler.ITaskHandler
}

func NewWorkerPoolProperties(taskHandler taskhandler.ITaskHandler) *WorkerPoolProperties {
	return &WorkerPoolProperties{
		taskTriggeringForProcessing: make(chan triggering.TaskTriggering),
		taskHandler:                 taskHandler}
}

type TaskExecutor struct {
	taskHandlerRegistry  handlerregistry.ITaskHandlerRegistry
	logger               *zap.Logger
	taskDao              dao.ITaskDao
	workerWg             *sync.WaitGroup
	workerPoolProperties []*WorkerPoolProperties
}

func NewTaskExecutor(taskHandlerRegistry handlerregistry.ITaskHandlerRegistry, taskDao dao.ITaskDao, logger *zap.Logger) *TaskExecutor {
	taskExecutor := &TaskExecutor{taskHandlerRegistry: taskHandlerRegistry, logger: logger, taskDao: taskDao, workerWg: &sync.WaitGroup{}}

	taskExecutor.launchWorkersPools(taskExecutor.workerWg)
	return taskExecutor
}

func (te *TaskExecutor) launchWorkersPools(workerWg *sync.WaitGroup) error {
	taskHandlers := te.taskHandlerRegistry.GetTaskHandlers()
	te.logger.Debug("Launching worker pools for task handlers.", zap.Int("task_handlers_count", len(taskHandlers)))

	for _, taskHandler := range taskHandlers {
		workerPoolProperties := NewWorkerPoolProperties(taskHandler)
		te.registerWorkerPoolProperties(workerPoolProperties)
		te.launchWorkerPool(workerPoolProperties, workerWg)
	}

	te.logger.Debug("Worker pools launched", zap.Int("worker_pool_count", len(te.workerPoolProperties)))

	return nil
}

func (te TaskExecutor) getWorkerPoolProperties(baseTask taskmodel.IBaseTask) (*WorkerPoolProperties, error) {
	var supportedWorkerPoolProperites []*WorkerPoolProperties
	for _, workerPoolProperties := range te.workerPoolProperties {
		if workerPoolProperties.taskHandler.Handles(baseTask) {
			supportedWorkerPoolProperites = append(supportedWorkerPoolProperites, workerPoolProperties)
		}
	}

	if len(supportedWorkerPoolProperites) == 0 {
		return nil, errors.New("No worker pool properties found")
	}

	if len(supportedWorkerPoolProperites) > 1 {
		return nil, errors.New("Multiple worker pool properties found")
	}

	return supportedWorkerPoolProperites[0], nil
}

func (te TaskExecutor) SubmitTask(task taskmodel.IBaseTask) error {
	te.logger.Debug("Submitting task", zap.String("task_id", task.GetId().String()))
	workerPoolProperties, err := te.getWorkerPoolProperties(task)
	if err != nil {
		te.logger.Error("Failed to get workerPoolProperties", zap.Error(err))
		return err
	}
	if workerPoolProperties == nil {
		te.logger.Error("WorkerPoolProperties is nil")
		return errors.New("WorkerPoolProperties is nil")
	}

	taskTriggering := triggering.NewTaskTriggering(task, time.Now().UTC())

	workerPoolProperties.taskTriggeringForProcessing <- *taskTriggering

	te.logger.Debug("Task submitted", zap.String("task_id", task.GetId().String()))

	return nil
}

func (te TaskExecutor) launchWorkerPool(workerPoolProperties *WorkerPoolProperties, workerWg *sync.WaitGroup) {
	te.logger.Debug("Launching worker pool")
	taskHandler := workerPoolProperties.taskHandler
	for {
		te.logger.Debug("Trying to book space for worker")
		bookspaceResponse := taskHandler.GetConcurrencyPolicy().BookSpace()
		if !bookspaceResponse.HasRoom() {
			te.logger.Debug("No room in worker pool")
			break
		}
		workerWg.Add(1)
		go te.launchWorker(workerPoolProperties, workerWg, func() {
			workerPoolProperties.taskHandler.GetConcurrencyPolicy().FreeSpace()
		})
	}
}

func (te *TaskExecutor) registerWorkerPoolProperties(workerPoolProperties *WorkerPoolProperties) {
	te.workerPoolProperties = append(te.workerPoolProperties, workerPoolProperties)
}

func (te TaskExecutor) launchWorker(workerPoolProperties *WorkerPoolProperties, workerWg *sync.WaitGroup, onFinish func()) error {
	defer onFinish()
	defer workerWg.Done()

	taskHandler := workerPoolProperties.taskHandler
	if taskHandler == nil {
		te.logger.Error("Task handler is nil")
		return errors.New("Task handler is nil")
	}

	for taskTriggering := range workerPoolProperties.taskTriggeringForProcessing {
		if !taskHandler.Handles(taskTriggering.GetTask()) {
			te.logger.Error("Task handler does not handle task", zap.Any("task", taskTriggering.GetTask()))
			continue
		}

		taskProcessor := taskHandler.GetProcessor(taskTriggering.GetTask())
		if taskProcessor == nil {
			te.logger.Error("Task processor is nil")
			continue
		}

		taskForProcessing, err := te.taskDao.GrabForProcessing(taskTriggering.GetTask(), time.Now().UTC().Add(time.Second*10))

		if err != nil {
			te.logger.Error("Failed to grab task for processing", zap.Error(err))
			te.markTaskAsError(taskTriggering.GetTask())
			continue
		}

		if taskForProcessing == nil {
			continue
		}

		processResultChan := make(chan taskprocessor.ProcessResult, 1)
		errChan := make(chan error, 1)

		processor := func() {
			processingResult, err := taskProcessor.Process(taskForProcessing)
			if err != nil {
				errChan <- err
				return
			}
			processResultChan <- processingResult
		}

		go processor()

		select {
		case processingResultHolder := <-processResultChan:
			te.handleProcessingResult(taskHandler, taskForProcessing, processingResultHolder)
		case processingErrorHolder := <-errChan:
			te.logger.Debug("Error while processing task", zap.Error(processingErrorHolder))
			te.setRetriesOrError(taskHandler, taskForProcessing)
		}

	}

	return nil
}

func (te TaskExecutor) handleProcessingResult(taskHandler taskhandler.ITaskHandler, task taskmodel.ITask, processingResult taskprocessor.ProcessResult) {
	if processingResult.GetResultCode() == taskprocessor.DONE {
		te.markTaskAsDone(task)
	} else if processingResult.GetResultCode() == taskprocessor.COMMIT_AND_RETRY {
		te.setRepeatOnSuccess(taskHandler, task)
	} else if processingResult.GetResultCode() == taskprocessor.DONE_AND_DELETE {
		te.deleteTask(task)
	}
}

func (te TaskExecutor) markTaskAsDone(t taskmodel.IBaseTask) bool {
	updated, err := te.taskDao.SetStatus(t.GetId(), taskmodel.DONE, t.GetVersion())
	if err != nil {
		te.logger.Error(err.Error())
		return false
	}
	return updated
}

func (s TaskExecutor) deleteTask(t taskmodel.IBaseTask) bool {
	updated, err := s.taskDao.DeleteTask(t.GetId(), t.GetVersion())
	if err != nil {
		s.logger.Error(err.Error())
		return false
	}
	return updated
}

func (s TaskExecutor) setRepeatOnSuccess(taskHandler taskhandler.ITaskHandler, task taskmodel.ITask) {
	retryPolicy := taskHandler.GetRetryPolicy(task)
	if retryPolicy.ResetTriesCountOnSuccess(task) {
		task.SetProcessingTriesCount(0)
	}
	retry, retryTime := retryPolicy.GetRetryTime(task)
	if !retry {
		s.setRetriesOrError(taskHandler, task)
	} else {
		s.logger.Info("Repeating task will be reprocessed", zap.Time("retry_time", retryTime), zap.Int("tries", task.GetProcessingTriesCount()), zap.Bool("reset_tries_count", retryPolicy.ResetTriesCountOnSuccess(task)), zap.Int("version", task.GetVersion()))
		s.setToBeRetried(task, retry, retryTime, retryPolicy.ResetTriesCountOnSuccess(task))
	}
}

func (te TaskExecutor) markTaskAsError(t taskmodel.IBaseTask) bool {
	updated, err := te.taskDao.SetStatus(t.GetId(), taskmodel.ERROR, t.GetVersion())
	if err != nil {
		te.logger.Error(err.Error())
		return false
	}
	return updated
}

func (te TaskExecutor) setRetriesOrError(taskHandler taskhandler.ITaskHandler, t taskmodel.ITask) {
	shouldRetry, retryTime := taskHandler.GetRetryPolicy(t).GetRetryTime(t)
	te.setRetriesOrError1(t, shouldRetry, retryTime)
}

func (te TaskExecutor) setRetriesOrError1(task taskmodel.ITask, shouldRetry bool, retryTime time.Time) {
	if !shouldRetry {
		te.logger.Info("Task marked as error", zap.String("task_id", task.GetId().String()))
		updated, err := te.taskDao.SetStatus(task.GetId(), taskmodel.ERROR, task.GetVersion())
		if err != nil {
			te.logger.Error(err.Error())
		}
		if !updated {
			te.logger.Error("Nothing was updated. Likely a version missmatch", zap.String("task_id", task.GetId().String()))
		}
	} else {
		te.setToBeRetried(task, shouldRetry, retryTime, false)
	}
}

func (te TaskExecutor) setToBeRetried(t taskmodel.ITask, shouldRetry bool, retryTime time.Time, resetTriesCount bool) {
	result, err := te.taskDao.SetToBeRetried(t.GetId(), shouldRetry, retryTime, t.GetVersion(), resetTriesCount)
	if err != nil {
		te.logger.Error(err.Error())
	}
	if !result {
		te.logger.Error("Nothing was updated. Likily a version missmatch", zap.String("task_id", t.GetId().String()))
		return
	}

	te.logger.Debug("Task set to be retried", zap.String("task_id", t.GetId().String()), zap.Time("retry_time", retryTime), zap.Bool("reset_tries_count", resetTriesCount))
}
