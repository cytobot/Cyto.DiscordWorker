package main

import (
	"fmt"
	"log"

	pbs "github.com/cytobot/messaging/transport/shared"
)

type WorkProcessor struct {
	client *DiscordClient
}

func NewWorkProcessor(discordClient *DiscordClient) *WorkProcessor {
	return &WorkProcessor{
		client: discordClient,
	}
}

func (p *WorkProcessor) processWork(req *pbs.DiscordWorkRequest) {
	msg := fmt.Sprintf("[Discord Work] Received work from %s for %s", req.SourceID, req.Command)
	log.Println(msg)

	//TODO: Handle commands
	/*
		err := p.client.SendMessage(req.ChannelID, msg)
		if err != nil {
			log.Println(err)
		}
	*/
}
