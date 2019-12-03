package main

import (
	"encoding/json"
	"runtime"
	"time"

	cytonats "github.com/cytobot/messaging/nats"
	pbd "github.com/cytobot/messaging/transport/discord"
	pbs "github.com/cytobot/messaging/transport/shared"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type NatsManager struct {
	client       *cytonats.NatsClient
	state        *workerState
	shutdownChan chan int32
}

var statsStartTime = time.Now()

func NewNatsManager(endpoint string, state *workerState) (*NatsManager, error) {
	client, err := cytonats.NewNatsClient(endpoint)
	if err != nil {
		return nil, err
	}

	return &NatsManager{
		client:       client,
		state:        state,
		shutdownChan: make(chan int32),
	}, nil
}

func (m *NatsManager) StartDiscordWorkListener(processor *WorkProcessor) error {
	discordWorkListener, err := m.client.ChanQueueSubscribe("discord_work", "discord-worker")
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-discordWorkListener:
				workRequest := &pbs.DiscordWorkRequest{}
				json.Unmarshal(msg.Data, workRequest)

				processor.processWork(workRequest)
			case <-m.shutdownChan:
				return
			}
		}
	}()

	return nil
}

func (m *NatsManager) StartHealthCheckInterval() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				sendHealthMessage(m)
			case <-m.shutdownChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (m *NatsManager) Shutdown() {
	m.shutdownChan <- 1
	m.client.Shutdown()
}

func sendHealthMessage(m *NatsManager) {
	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)

	content := &pbd.HealthCheckStatus{
		Timestamp:     mapToProtoTimestamp(time.Now().UTC()),
		InstanceID:    m.state.id,
		ShardID:       int32(m.state.shardID),
		Uptime:        time.Now().Sub(statsStartTime).Nanoseconds(),
		MemAllocated:  int64(stats.Alloc),
		MemSystem:     int64(stats.Sys),
		MemCumulative: int64(stats.TotalAlloc),
		TaskCount:     int32(runtime.NumGoroutine()),
	}

	m.client.Publish("worker_health", content)
}

func mapToProtoTimestamp(timeValue time.Time) *timestamp.Timestamp {
	protoTimestamp, _ := ptypes.TimestampProto(timeValue)
	return protoTimestamp
}
