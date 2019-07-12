package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nsyszr/lcm/pkg/devicecontrol/controlchannel/message"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatal("missing argument cli command")
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	type m3CliCommand struct {
		Command string `json:"command"`
	}

	req := message.CallRequest{
		TargetType: message.TargetTypeDevice,
		TargetID:   os.Args[1],
		Command:    "m3_cli",
		Arguments: m3CliCommand{
			Command: os.Args[2],
		},
	}
	fmt.Printf("req=%v", req)
	requestData, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	replyMsg, err := nc.Request("iotcore.devicecontrol.v1.default.call", requestData, 16*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	rep := message.CallReply{}
	if err := json.Unmarshal(replyMsg.Data, &rep); err != nil {
		log.Fatal(err)
	}

	if rep.Status == message.ReplyStatusSuccess {
		j, _ := json.Marshal(rep.Results)
		fmt.Printf("%s", j)
	} else {
		fmt.Printf("error:\n %s\n\n", rep.ErrorReason)
	}
}
