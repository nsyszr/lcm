package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	// Subscribe
	if _, err := nc.Subscribe("iotcore.devicecontrol.v1.default.events.>", func(m *nats.Msg) {
		fmt.Printf("subject: %s, message: %s\n", m.Subject, string(m.Data))
	}); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quitCh := make(chan os.Signal)
	signal.Notify(quitCh, os.Interrupt)
	<-quitCh
}
