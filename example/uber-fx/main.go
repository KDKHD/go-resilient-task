package main

import (
	taskhandler "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-core/pkg/handler/task_handler"
	"go.uber.org/fx"
)

func main() {
	fx.New(
		fx.Provide(
			NewLogger,
			NewKgoClient,
			NewPostgresClient,
			NewTaskProperties,
			NewResilientTaskConfiguration,
			AsHandler(PaymentInitiatedHandler),
			AsHandler(PaymentProcessedHandler),
		),
		InitiateResilientTaskFx(),
	).Run()
}

func AsHandler(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(taskhandler.ITaskHandler)),
		fx.ResultTags(`group:"handlers"`),
	)
}
