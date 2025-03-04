### Sample task handler
```go
taskhandleradapter.NewTaskHandlerAdapterBuilder(
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
```