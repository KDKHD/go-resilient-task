package main

import (
	autoconfiguration "github.com/KDKHD/go-resilient-task/modules/go-resilient-task-uber-fx/auto-configuration"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		autoconfiguration.Provider(),
		fx.Decorate(func() (*zap.Logger, error) {
			return zap.NewDevelopment()
		}),
		fx.Provide(
			autoconfiguration.AsHandler(PaymentInitiatedHandler),
			autoconfiguration.AsHandler(PaymentProcessedHandler),
			NewKgoClient,
			NewPostgresClient,
			NewTaskProperties,
		),
		autoconfiguration.InitiateResilientTaskFx(),
	).Run()
}
