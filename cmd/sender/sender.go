package main

import (
	"fmt"
	"strconv"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"lucathurm.dev/tofuh/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	options := mqtt.NewClientOptions()
	options.ClientID = "tofuh-test"
	options.AddBroker("tcp://localhost:1883")
	client := mqtt.NewClient(options)
	if tkn := client.Connect(); tkn.WaitTimeout(cfg.MsgConnTimeout) && tkn.Error() != nil {
		panic(tkn.Error())
	}

	client.Subscribe(cfg.MsgStateTopicKey, cfg.MsgQOS, func(c mqtt.Client, m mqtt.Message) {
		println("new state: %64b")
	})

	var input string
	for {
		_, iErr := fmt.Scanln(&input)
		if iErr != nil {
			print("failed to read line")
			continue
		}
		b, parseErr := strconv.ParseUint(input, 2, 8)
		if parseErr != nil {
			print("failed to parse line")
			continue
		}
		pubTkn := client.Publish(cfg.MsgSetBitTopicKey, cfg.MsgQOS, false, []byte{byte(b)})
		if pubTkn.Error() != nil {
			fmt.Println(pubTkn.Error())
			print("failed to publish message")
		}
	}
}
