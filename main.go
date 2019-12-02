package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
)

type worker struct {
	discord       *DiscordClient
	nats          *NatsManager
	workProcessor *WorkProcessor
}

func main() {
	discordClient := getDiscordClient()

	worker := &worker{
		discord:       discordClient,
		nats:          getNatsManager(),
		workProcessor: NewWorkProcessor(discordClient),
	}

	go worker.nats.StartDiscordWorkListener(worker.workProcessor)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

out:
	for {
		select {
		case <-c:
			worker.nats.Shutdown()
			break out
		}
	}
}

func getDiscordClient() *DiscordClient {
	token := os.Getenv("DiscordToken")

	if token == "" {
		panic("No token provided.")
	}

	ownerUserID := os.Getenv("DiscordOwnerId")

	args := []interface{}{("Bot " + token)}

	client := &DiscordClient{
		args:        args,
		OwnerUserID: ownerUserID,
	}

	err := client.Open()
	if err != nil {
		panic(fmt.Sprintf("[Discord error] %s", err))
	}

	return client
}

func getNatsManager() *NatsManager {
	natsEndpoint := os.Getenv("NatsEndpoint")

	if natsEndpoint == "" {
		panic("No nats endpoint provided.")
	}

	client, err := NewNatsManager(natsEndpoint)
	if err != nil {
		panic(fmt.Sprintf("[NATS error] %s", err))
	}

	log.Println("Connected to NATS")

	return client
}
