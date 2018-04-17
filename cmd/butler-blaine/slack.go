package main

import (
	"io/ioutil"
	"log"

	"github.com/megasay/checkerr"
	"github.com/nlopes/slack"
)

func startSlackBot() {
	log.Printf("INFO:\tStarting Slack bot.\n")
	token, err := ioutil.ReadFile(*flagSlackToken)
	checkerr.Check(err, checkerr.ERROR, "Unable to read slack bot token file.")
	api := slack.New(string(token))

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		log.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			log.Println("Infos:", ev.Info)
			log.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			//rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", "C2147483705"))

		case *slack.MessageEvent:
			log.Printf("Message: %v\n", ev.Msg.Text)
			answer := getResponse(ev.Msg.Text)
			log.Println(answer)
			rtm.SendMessage(rtm.NewOutgoingMessage(answer, ev.Channel))

		case *slack.PresenceChangeEvent:
			log.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			log.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			log.Printf("Invalid credentials")
			return

		default:
			// Ignore other events..
			//log.Printf("Unexpected: %v\n", msg.Data)
		}
	}
}
