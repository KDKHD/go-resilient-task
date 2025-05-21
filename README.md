# Go Resilient Task

A highly reliable distributed task execution framework written in Go, designed to ensure that triggered tasks will eventually execute, even in the face of system failures or network issues.

## Overview

Go Resilient Task provides a robust infrastructure for scheduling, executing, and monitoring distributed tasks with built-in reliability guarantees. It's designed for systems where task execution reliability is critical, such as payment processing, order fulfillment, or any workflow where tasks must complete despite temporary failures.

## Key Features

- **Guaranteed Task Execution**: Once a task is triggered, the framework ensures it will eventually execute to completion
- **Flexible Retry Policies**: Configurable retry strategies including exponential backoff with customizable parameters
- **Concurrency Control**: Limit the number of concurrent task executions to prevent system overload
- **Processing Policies**: Define how long tasks can run before timing out
- **Task Scheduling**: Schedule tasks to run at specific times in the future
- **Task Deduplication**: Prevent duplicate task execution using unique keys
- **Kafka Integration**: Use Kafka for reliable task distribution and execution triggering
- **PostgreSQL Persistence**: Store task state in PostgreSQL for durability
- **Uber-FX Integration**: Easy integration with the Uber-FX dependency injection framework
- **Stuck Task Detection**: Identify and recover from tasks that have been processing for too long

## Architecture

The framework consists of several key components:

- **Task Service**: Manages task creation, scheduling, and status tracking
- **Task Executor**: Executes tasks using registered task handlers
- **Task Handlers**: Process specific task types with custom business logic
- **Retry Policies**: Define how and when to retry failed tasks
- **Concurrency Policies**: Control how many tasks can execute simultaneously
- **Processing Policies**: Define task execution timeouts
- **Task DAO**: Persistence layer for storing task state
- **Task Execution Triggerer**: Triggers task execution based on events (e.g., Kafka messages)

## Getting Started

### Prerequisites

- Go 1.18 or higher
- PostgreSQL
- Kafka (optional, for distributed task execution)

### Installation

```bash
go get github.com/KDKHD/go-resilient-task
```

### Basic Usage

1. Define your task handlers:

```go
func PaymentProcessedHandler(logger *zap.Logger) taskhandler.ITaskHandler {
    return taskhandleradapter.
        NewTaskHandlerAdapterBuilder(
            func(task taskmodel.IBaseTask) bool {
                return task.GetType() == "payment_processed"
            },
            taskprocessor.TaskProcessorFunc(
                func(task taskmodel.ITask) (taskprocessor.ProcessResult, error) {
                    logger.Debug("processing payment", zap.String("task_data", task.GetData()))
                    // Your business logic here
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
}
```

2. Configure and start the framework with Uber-FX:

```go
fx.New(
    autoconfiguration.Provider(), // Default configuration
    fx.Provide(
        autoconfiguration.AsHandler(PaymentProcessedHandler), // Register handlers
        NewPostgresClient,
        NewTaskProperties,
    ),
    autoconfiguration.InitiateResilientTaskFx(), // Start processing
).Run()
```

3. Add tasks to be executed:

```go
taskService.AddTask(taskservice.AddTaskRequest{
    Type: "payment_processed",
    TaskId: uuid.New(),
    Data: []byte(`{"orderId": "12345"}`),
    RunAfterTime: time.Now().UTC().Add(time.Second * 10),
    ExpectedQueueTime: time.Second * 120,
})
```

## Configuration Options

### Task Properties

```go
taskproperties.NewTaskProperties(
    taskproperties.WithTaskStuckTimeout(time.Minute*20),
    taskproperties.WithTaskResumer(
        taskproperties.NewTaskResumerProperties(
            taskproperties.WithBatchSize(1000),
            taskproperties.WithPollingInterval(time.Second*2),
            taskproperties.WithConcurrency(10),
        )),
    taskproperties.WithKafka(
        taskproperties.NewKafkaProperties(
            taskproperties.WithBootstrapServers("localhost:29092"),
            taskproperties.WithKafkaTopicsNamespace("my-namespace"),
            taskproperties.WithKafkaConsumerGroupId("tasks"),
        )),
)
```

### Retry Policies

- **Exponential Retry Policy**: Increases delay between retries exponentially
- **No Retry Policy**: Doesn't retry failed tasks

### Concurrency Policies

- **Simple Concurrency Policy**: Limits the number of concurrent task executions

### Processing Policies

- **Simple Processing Policy**: Sets a timeout for task execution

## Example Project

Check out the example in the `example/uber-fx` directory for a complete working example using Uber-FX, PostgreSQL, and Kafka.

To run the example:

1. Start the required infrastructure:
   ```bash
   cd example/uber-fx
   docker compose up
   ```

2. Run the example application:
   ```bash
   go run .
   ```

## Contributing

We welcome contributions to the Go Resilient Task project! Here's how you can help:

### Reporting Issues

- Use the GitHub issue tracker to report bugs
- Describe the bug or feature request in detail
- Include code examples and reproduction steps

### Development Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Coding Standards

- Follow Go best practices and style guidelines
- Write tests for new features
- Keep the code modular and maintainable
- Document public APIs

### Running Tests

```bash
go test ./...
```

## License

[MIT License](LICENSE)
