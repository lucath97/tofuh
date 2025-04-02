package core

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func Sender() {
	options := mqtt.NewClientOptions()
	options.AddBroker("tcp://localhost:1883")
	options.SetClientID("mysender")

	options.OnConnect = func(c mqtt.Client) {
		fmt.Println("Sender connected")
		t := c.Publish("topic/t1", 1, true, make([]byte, 1))
		t.WaitTimeout(time.Second * 5)
		c.Disconnect(1000)
		fmt.Println("disconnected")
	}

	client := mqtt.NewClient(options)
	if token := client.Connect(); token.WaitTimeout(time.Second*5) && token.Error() != nil {
		panic(token.Error())
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
}
