package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	cytonats "github.com/cytobot/messaging/nats"
	"github.com/lampjaw/discordclient"
	"github.com/lithammer/shortuuid"
)

type workerState struct {
	id            string
	shardID       int
	discord       *discordclient.DiscordClient
	nats          *NatsManager
	workProcessor *WorkProcessor
}

func main() {
	discordClient := getDiscordClient()

	worker := &workerState{
		id:      shortuuid.New(),
		shardID: getShardID(),
		discord: discordClient,
	}

	worker.nats = getNatsManager(worker)
	worker.workProcessor = getWorkProcessor(discordClient, worker.nats.client)

	go worker.nats.StartDiscordWorkListener(worker.workProcessor)
	go worker.nats.StartHealthCheckInterval()

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

func getShardID() int {
	envShardID := os.Getenv("ShardId")
	if envShardID != "" {
		shardID, _ := strconv.ParseInt(envShardID, 10, 64)
		return int(shardID)
	}
	return -1
}

func getDiscordClient() *discordclient.DiscordClient {
	token := os.Getenv("DiscordToken")

	if token == "" {
		panic("No token provided.")
	}

	clientID := os.Getenv("DiscordClientId")
	ownerUserID := os.Getenv("DiscordOwnerId")

	client := discordclient.NewDiscordClient(token, ownerUserID, clientID)

	err := client.Open()
	if err != nil {
		panic(fmt.Sprintf("[Discord error] %s", err))
	}

	return client
}

func getNatsManager(s *workerState) *NatsManager {
	natsEndpoint := os.Getenv("NatsEndpoint")

	if natsEndpoint == "" {
		panic("No nats endpoint provided.")
	}

	client, err := NewNatsManager(natsEndpoint, s)
	if err != nil {
		panic(fmt.Sprintf("[NATS error] %s", err))
	}

	log.Println("Connected to NATS")

	return client
}

func getWorkProcessor(discordClient *discordclient.DiscordClient, natsClient *cytonats.NatsClient) *WorkProcessor {
	managerEndpoint := os.Getenv("ManagerEndpoint")

	if managerEndpoint == "" {
		panic("No manager endpoint provided.")
	}

	processor := NewWorkProcessor(discordClient, natsClient, managerEndpoint)

	return processor
}
