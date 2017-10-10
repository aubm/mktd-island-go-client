package main

import (
	"context"
	"os"
	"os/signal"

	"mktd5/mktd-island/client/game"
	"mktd5/mktd-island/client/game/strategy"
	"mktd5/mktd-island/client/log"
	"mktd5/mktd-island/client/mediator"
	"mktd5/mktd-island/client/utils"

	"github.com/facebookgo/inject"
	"github.com/jessevdk/go-flags"
)

var (
	appConfig           = &utils.AppConfig{}
	gameAgent           = &game.Agent{}
	defaultMoveStrategy = &strategy.DefaultMoveStrategy{}
	mediatorClient      = &mediator.Client{}
	logger              = &log.Logger{}
)

func main() {
	if _, err := flags.Parse(appConfig); err != nil {
		exit(err)
	}

	if err := inject.Populate(
		appConfig,
		gameAgent,
		defaultMoveStrategy,
		mediatorClient,
		logger,
	); err != nil {
		exit(err)
	}

	if len(appConfig.Verbose) > 0 {
		log.ConfigureDebugLevel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	gameAgent.StopCallback = cancel
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	go func() {
		for range signals {
			logger.Info("received interrupt signal", nil)
			cancel()
		}
	}()

	if err := gameAgent.StartClientServer(ctx); err != nil {
		exit(err)
	}

	logger.Info("bye", nil)
}

func exit(err error) {
	logger.Fatal(err.Error(), nil)
}
