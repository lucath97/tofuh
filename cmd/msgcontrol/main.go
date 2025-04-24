package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/redis/go-redis/v9"
	"lucathurm.dev/tofuh/internal/config"
	"lucathurm.dev/tofuh/internal/db"
	"lucathurm.dev/tofuh/internal/msg"
)

const (
	channelFmt = "__keyspace@%d__:%s"
)

func main() {
	log := slog.Default()

	cfg := config.LoadConfig()
	log.Info("loaded config")

	msgOptions := mqtt.NewClientOptions()
	msgOptions.AddBroker(cfg.MsgAddress)
	msgOptions.ClientID = cfg.MsgClientID
	msgOptions.SetAutoReconnect(true)
	client := mqtt.NewClient(msgOptions)
	log.Info("created mqtt client")

	if token := client.Connect(); token.WaitTimeout(cfg.MsgConnTimeout) && token.Error() != nil {
		log.Error("failed to connect to mqtt broker")
		panic(token.Error())
	}
	defer client.Disconnect(uint(cfg.MsgConnTimeout))
	log.Info("connected to mqtt broker")

	rds := redis.NewClient(&redis.Options{Addr: cfg.DbAddress, Password: cfg.DbPassword})
	log.Info("created redis client")

	pingCtx, pingCtxCancel := context.WithTimeout(context.Background(), cfg.DbTimeout)
	if pingErr := rds.Ping(pingCtx).Err(); pingErr != nil {
		log.Error("failed to connect to database", "error", pingErr)
		panic(pingErr)
	}
	pingCtxCancel()
	log.Info("connected to database")

	rdsSub := rds.Subscribe(context.Background(), fmt.Sprintf(channelFmt, 0, cfg.DbStateKey))
	defer rdsSub.Close()

	go func() {
		for message := range rdsSub.Channel() {
			dbCtx, dbCtxCancel := context.WithTimeout(context.Background(), cfg.DbTimeout)
			log.Info("received state update", message.Channel, message.Payload)

			state, getErr := db.GetState(rds, dbCtx, cfg.DbStateKey)
			dbCtxCancel()
			if getErr != nil {
				log.Error("failed to get state from database", "error", getErr)
				continue
			}
			log.Info("retrieved new state from database", "state", state)

			pubErr := msg.Publish(&client, cfg.MsgStateTopicKey, cfg.MsgQOS, cfg.MsgRetainState, state[:])
			if pubErr != nil {
				log.Error("failed to publish state to message broker", "error", pubErr)
				continue
			}
			log.Info("published new state to message broker")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("shutting down")
}
