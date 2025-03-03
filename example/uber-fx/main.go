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
}
