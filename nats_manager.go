package main

import (
	"encoding/json"

	cytonats "github.com/cytobot/messaging/nats"
	pbs "github.com/cytobot/messaging/transport/shared"
)

type NatsManager struct {
	client       *cytonats.NatsClient
	listenerChan chan int32
}

func NewNatsManager(endpoint string) (*NatsManager, error) {
	client, err := cytonats.NewNatsClient(endpoint)
	if err != nil {
		return nil, err
	}

	return &NatsManager{
		client: client,
	}, nil
}

func (m *NatsManager) StartDiscordWorkListener(processor *WorkProcessor) error {
	discordWorkListener, err := m.client.ChanQueueSubscribe("discord_work", "discord-worker")
	if err != nil {
		return err
	}

	m.listenerChan = make(chan int32)
	go func() {
		for {
			select {
			case msg := <-discordWorkListener:
				workRequest := &pbs.DiscordWorkRequest{}
				json.Unmarshal(msg.Data, workRequest)

				go processor.processWork(workRequest)
			case <-m.listenerChan:
				return
			}
		}
	}()

	return nil
}

func (m *NatsManager) Shutdown() {
	if m.listenerChan != nil {
		m.listenerChan <- 0
	}
	m.client.Shutdown()
}
