package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Client() {
	options := mqtt.NewClientOptions()
	options.AddBroker("tcp://localhost:1883")
	options.SetClientID("myclient")

	options.OnConnect = func(c mqtt.Client) {
		print("connection established")
		c.Subscribe("topic/t1", 1, func(c mqtt.Client, m mqtt.Message) {
			fmt.Println(m.Payload())
			m.Ack()
		})
	}

	client := mqtt.NewClient(options)

	if token := client.Connect(); token.WaitTimeout(time.Second*5) && token.Error() != nil {
		panic(token.Error())
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	client.Unsubscribe("topic/t1")
	client.Disconnect(0)
}
