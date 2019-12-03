package main

import (
	"log"

	"github.com/cytobot/commandworks"
	cytonats "github.com/cytobot/messaging/nats"
	pbs "github.com/cytobot/messaging/transport/shared"
	"github.com/lampjaw/discordclient"
)

type WorkProcessor struct {
	client         *discordclient.DiscordClient
	commandHandler *commandworks.CommandHandler
}

func NewWorkProcessor(discordClient *discordclient.DiscordClient, natsClient *cytonats.NatsClient, managerEndpoint string) *WorkProcessor {
	return &WorkProcessor{
		client:         discordClient,
		commandHandler: commandworks.NewCommandHandler(managerEndpoint, natsClient),
	}
}

func (p *WorkProcessor) processWork(req *pbs.DiscordWorkRequest) {
	log.Printf("[Discord Work] Received work from %s for %s (%s)", req.SourceID, req.Command, req.MessageID)

	var processErr error
	if req.Type == "command" {
		processErr = p.commandHandler.ProcessCommand(p.client, req)
	}

	if processErr != nil {
		log.Printf("[WorkProcessor] Failed to process work: %s", processErr)
	}
}
