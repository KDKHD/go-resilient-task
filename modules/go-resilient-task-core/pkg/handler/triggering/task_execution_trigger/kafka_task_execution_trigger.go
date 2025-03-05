package taskexecutiontrigger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/dao"
	handlerregistry "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/handler_registry"
	taskexecutor "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_executor"
	taskmodel "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/model/task"
	"github.com/pkg/errors"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type KafkaTaskExecutionTrigger struct {
	taskHandlerRegistry handlerregistry.ITaskHandlerRegistry
	lifecycleLock       *sync.Mutex
	logger              *zap.Logger
	kafkaClient         *kgo.Client
	taskExecutor        taskexecutor.ITaskExecutor
	taskDao             dao.ITaskDao
}

func NewKafkaTaskExecutionTrigger(taskHandlerRegistry handlerregistry.ITaskHandlerRegistry, logger *zap.Logger, kafkaClient *kgo.Client, taskExecutor taskexecutor.ITaskExecutor, taskDao dao.ITaskDao) *KafkaTaskExecutionTrigger {
	fmt.Print("taskHandlerRegistry")
	fmt.Print(taskHandlerRegistry)

	return &KafkaTaskExecutionTrigger{
		taskHandlerRegistry: taskHandlerRegistry,
		lifecycleLock:       &sync.Mutex{},
		logger:              logger,
		kafkaClient:         kafkaClient,
		taskExecutor:        taskExecutor,
		taskDao:             taskDao,
	}
}

func (ket *KafkaTaskExecutionTrigger) Trigger(task taskmodel.IBaseTask) error {
	ket.logger.Debug("Triggering task", zap.String("task_id", task.GetId().String()))
	taskHandler, err := ket.taskHandlerRegistry.GetTaskHandler(task)
	if err != nil {
		ket.logger.Error("Failed to get task handler", zap.Error(err))
		result, err := ket.taskDao.SetStatus(task.GetId(), taskmodel.ERROR, task.GetVersion())
		if !result || err != nil {
			ket.logger.Error("Failed to set task status", zap.Error(err))
		}
		return err
	}

	if taskHandler == nil {
		ket.logger.Error("Task handler not found")
		return errors.New("Task handler not found")
	}

	taskSt, err := json.Marshal(task)

	if err != nil {
		ket.logger.Error("Failed to marshal task", zap.Error(err))
		return err
	}

	record := &kgo.Record{
		Topic:     "tasks",
		Value:     taskSt,
		Timestamp: time.Now().UTC(),
	}

	ket.kafkaClient.Produce(context.Background(), record, func(_ *kgo.Record, err error) {
		if err != nil {
			ket.logger.Error("record had a produce error", zap.Error(err))
		}
	})

	return nil
}

func (ket *KafkaTaskExecutionTrigger) poll() {

	for {
		fetches := ket.kafkaClient.PollFetches(context.Background())

		maxTopicPartitionOffsetRecord := make(map[string]map[int32]*kgo.Record)
		ket.logger.Debug("Fetched records", zap.Int("count", len(fetches.Records())))
		for _, record := range fetches.Records() {
			var basetask *taskmodel.BaseTask
			error := json.Unmarshal(record.Value, &basetask)
			if error != nil {
				ket.logger.Error("Failed to unmarshal task", zap.Error(error))
				continue
			}

			ket.taskExecutor.SubmitTask(basetask)

			if _, ok := maxTopicPartitionOffsetRecord[record.Topic]; !ok {
				maxTopicPartitionOffsetRecord[record.Topic] = make(map[int32]*kgo.Record)
			}

			if _, ok := maxTopicPartitionOffsetRecord[record.Topic][record.Partition]; !ok {
				maxTopicPartitionOffsetRecord[record.Topic][record.Partition] = record
			} else {
				if maxTopicPartitionOffsetRecord[record.Topic][record.Partition].Offset < record.Offset {
					maxTopicPartitionOffsetRecord[record.Topic][record.Partition] = record
				}
			}

		}

		rs := make([]*kgo.Record, 0)
		for _, records := range maxTopicPartitionOffsetRecord {
			for _, record := range records {
				rs = append(rs, record)
			}
		}

		if err := ket.kafkaClient.CommitRecords(context.Background(), rs...); err != nil {
			ket.logger.Error("commit records failed", zap.Error(err))
			continue
		}
	}
}

func (ket *KafkaTaskExecutionTrigger) StartTasksProcessing() error {
	go ket.poll()
	return nil
}

func (ket *KafkaTaskExecutionTrigger) StopTasksProcessing() error {
	return nil
}
