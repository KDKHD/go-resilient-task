```go
logger, _ := zap.NewDevelopment()
defer logger.Sync() // Flushes buffer, if any

kafkaClient, err := s.createKafkaClient(context.Background(),
    kgo.ConsumerGroup("my-group-identifier"),
    kgo.ConsumeTopics("tasks"))

if err != nil {
    logger.Error("Failed to create kafka client", zap.Error(err))
}

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
)

configuration.GetTaskResumer().ResumeProcessing()
configuration.GetTaskExecutionTrigger().StartTasksProcessing()

for i := 0; i < 100; i++ {
    addTaskResponse, error := configuration.GetTaskService().AddTask(taskservice.AddTaskRequest{Type: "test", TaskId: uuid.New(), Data: []byte(fmt.Sprintf("TaskNumber %d", i)), RunAfterTime: time.Now().UTC(), ExpectedQueueTime: time.Second * 120})

    s.Require().NoError(error)

    assert.NotNil(s.T(), addTaskResponse, "addTaskResponse should not be nil")
    assert.NotNil(s.T(), addTaskResponse.TaskId.String(), "addTaskResponse.TaskId should not be nil")
}

logger.Debug("Starting tasks processing")
```