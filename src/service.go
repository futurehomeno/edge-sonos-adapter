package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/futurehomeno/fimpgo"
	"github.com/futurehomeno/fimpgo/discovery"
	log "github.com/sirupsen/logrus"
	"github.com/thingsplex/sonos/model"
	"github.com/thingsplex/sonos/router"
	"github.com/thingsplex/sonos/utils"
)

func main() {
	var workDir string
	flag.StringVar(&workDir, "c", "", "Work dir")
	flag.Parse()
	if workDir == "" {
		workDir = "./"
	} else {
		fmt.Println("Work dir ", workDir)
	}
	appLifecycle := model.NewAppLifecycle()
	configs := model.NewConfigs(workDir)
	err := configs.LoadFromFile()
	if err != nil {
		fmt.Print(err)
		panic("Can't load config file.")
	}

	utils.SetupLog(configs.LogFile, configs.LogLevel, configs.LogFormat)
	log.Info("--------------Starting sonos----------------")
	log.Info("Work directory : ", configs.WorkDir)
	appLifecycle.PublishEvent(model.EventConfiguring, "main", nil)

	mqtt := fimpgo.NewMqttTransport(configs.MqttServerURI, configs.MqttClientIdPrefix, configs.MqttUsername, configs.MqttPassword, true, 1, 1)
	err = mqtt.Start()
	responder := discovery.NewServiceDiscoveryResponder(mqtt)
	responder.RegisterResource(model.GetDiscoveryResource())
	responder.Start()

	fimpRouter := router.NewFromFimpRouter(mqtt, appLifecycle, configs)
	fimpRouter.Start()

	appLifecycle.SetConnectionState(model.ConnStateDisconnected)
	if configs.IsConfigured() && err == nil {
		appLifecycle.SetConfigState(model.ConfigStateConfigured)
		appLifecycle.SetAppState(model.AppStateRunning, nil)
		appLifecycle.SetConnectionState(model.ConnStateConnected)
	} else {
		appLifecycle.SetConfigState(model.ConfigStateNotConfigured)
		appLifecycle.SetAppState(model.AppStateNotConfigured, nil)
		appLifecycle.SetConnectionState(model.ConnStateDisconnected)
	}

	if configs.IsAuthenticated() && err == nil {
		appLifecycle.SetAuthState(model.AuthStateAuthenticated)
	} else {
		appLifecycle.SetAuthState(model.AuthStateNotAuthenticated)
	}

	for {
		appLifecycle.WaitForState("main", model.AppStateRunning)
		appLifecycle.WaitForState(model.AppStateNotConfigured, "main")
	}

	mqtt.Stop()
	time.Sleep(5 * time.Second)
}
