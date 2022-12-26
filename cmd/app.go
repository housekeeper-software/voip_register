package cmd

import (
	"context"
	"github.com/sirupsen/logrus"
	"jingxi.cn/voip_register/conf"
	"jingxi.cn/voip_register/controller"
	"os"
	"os/signal"
	"syscall"
)

type App struct {
	ctx        context.Context
	cancel     context.CancelFunc
	controller *controller.Controller
}

func NewApp() *App {
	return &App{
		ctx:        nil,
		cancel:     nil,
		controller: nil,
	}
}

func (app *App) Run(httpAddr string, serverConf* conf.ServerConfig) {
	app.ctx, app.cancel = context.WithCancel(context.Background())
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		app.Quit()
		logrus.Infof("server gracefully shutdown")
	}()

	app.controller = controller.NewController(serverConf)
	go func() {
		if err := app.controller.Run(httpAddr); nil != err {
			logrus.Errorf("server run failed, err: %+v", err)
			app.Quit()
		}
	}()

Loop:
	for {
		select {
		case <-app.ctx.Done():
			break Loop
		}
	}
	app.controller.Stop()
	logrus.Infof("app quit!")
}

func (app *App) Quit() {
	app.cancel()
}
