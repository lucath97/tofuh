package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/redis/go-redis/v9"
	"lucathurm.dev/tofuh/internal/config"
	"lucathurm.dev/tofuh/internal/db"
	"lucathurm.dev/tofuh/internal/msg"
)

func main() {
	log := slog.Default()

	cfg := config.LoadConfig()
	log.Info("loaded config")

	rds := redis.NewClient(&redis.Options{Addr: cfg.DbAddress, Password: cfg.DbPassword})
	log.Info("created redis client")

	dbCtx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	msgOptions := mqtt.NewClientOptions()
	msgOptions.AddBroker(cfg.MsgAddress)
	msgOptions.ClientID = cfg.MsgClientID
	client := mqtt.NewClient(msgOptions)
	log.Info("created mqtt client")

	if token := client.Connect(); token.WaitTimeout(cfg.MsgConnTimeout) && token.Error() != nil {
		log.Error("failed to connect to mqtt broker")
		panic(token.Error())
	}
	log.Info("connected to mqtt broker")

	msg.Subscribe(&client, cfg.MsgSetBitTopicKey, cfg.MsgQOS, func(c mqtt.Client, m mqtt.Message) {
		payload := m.Payload()
		if len(payload) != 1 {
			log.Error("received malformed message")
			return
		}

		pos, set := msg.SetBitMsg(payload).Unmarshal()

		setErr := db.SetBit(rds, dbCtx, cfg.DbStateKey, pos, set)
		if setErr != nil {
			log.Error(setErr.Error())
			log.Error("failed to set bit")
			return
		}
		log.Info("set bit")

		newState, getErr := db.GetState(rds, dbCtx, cfg.DbStateKey)
		if getErr != nil {
			log.Error(getErr.Error())
			log.Error("failed to get state")
			return
		}
		log.Info("read new state")

		pubErr := msg.Publish(&client, cfg.MsgStateTopicKey, cfg.MsgQOS, cfg.MsgRetainState, newState[:])
		if pubErr != nil {
			log.Error(pubErr.Error())
			log.Error("failed to publish new state")
		}
		log.Info("published new state")
	})

	log.Info("listening for SIGINT / SIGTERM to shutdown broker")
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
